package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
	"url-short/internal/models"
	"url-short/internal/repositories"
	"url-short/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

type LinkHandler struct {
	LinkRepo     *repositories.LinkRepository
	AnalyticRepo *repositories.AnalyticRepository
}

type IPGeoResponse struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

var locationCache sync.Map

func getLocationWithCache(ip string) (string, error) {
	if cached, ok := locationCache.Load(ip); ok {
		return cached.(string), nil
	}

	location, err := getLocation(ip)
	if err != nil {
		return "", err
	}

	locationCache.Store(ip, location)
	time.AfterFunc(24*time.Hour, func() {
		locationCache.Delete(ip)
	})

	return location, nil
}

func getLocation(ip string) (string, error) {
	if ip == "::1" || ip == "127.0.0.1" {
		return "localhost", nil
	}

	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API вернуло статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var geoData IPGeoResponse
	if err := json.Unmarshal(body, &geoData); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	if geoData.City == "" && geoData.Country == "" {
		return "unknown", nil
	}

	return geoData.City + ", " + geoData.Country, nil
}

func isValidCustomCode(code string) bool {
	if len(code) < 2 || len(code) > 20 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(code)
}

func (h *LinkHandler) CreateShortLink(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не аутентифицирован"})
		return
	}

	var req struct {
		OriginalURL string `json:"original_url" binding:"required,url"`
		CustomCode  string `json:"custom_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный URL"})
		return
	}

	var shortCode string
	var err error

	if req.CustomCode != "" {
		if !isValidCustomCode(req.CustomCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Код должен содержать 2-20 символов (a-z, A-Z, 0-9, _, -)"})
			return
		}

		exists, err := h.LinkRepo.IsShortCodeExist(req.CustomCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки кода"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Код уже занят"})
			return
		}
		shortCode = req.CustomCode
	} else {
		shortCode, err = utils.GenerateUniqueShortCode(h.LinkRepo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации кода"})
			return
		}
	}

	link := &models.Link{
		UserID:      userID.(int),
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
	}

	if err := h.LinkRepo.CreateLink(link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ссылки"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_code": shortCode})
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код не указан"})
		return
	}

	link, err := h.LinkRepo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, repositories.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}

	location, err := getLocationWithCache(c.ClientIP())
	if err != nil {
		log.Printf("Геолокация для IP %s не удалась: %v", c.ClientIP(), err)
		location = "unknown"
	}

	if err := h.LinkRepo.IncrementClickCount(shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления счетчика"})
		return
	}

	ua := useragent.Parse(c.GetHeader("User-Agent"))

	clickData := &models.ClickAnalytic{
		LinkID:     link.ID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Location:   location,
		DeviceType: ua.Device,
		OS:         ua.OS,
		Browser:    ua.Name,
		ClickedAt:  time.Now(),
	}

	if err := h.AnalyticRepo.SaveClick(clickData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аналитики"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
}

func (h *LinkHandler) GetLinkStats(c *gin.Context) {
	shortCode := c.Param("short_code")
	link, err := h.LinkRepo.FindByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
		return
	}

	stats, err := h.AnalyticRepo.GetAnalytics(link.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
