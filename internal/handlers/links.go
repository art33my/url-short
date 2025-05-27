package handlers

import (
	"encoding/json"
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

// CreateShortLink godoc
// @Summary Создать короткую ссылку
// @Tags links
// @Security ApiKeyAuth
// @Accept  json
// @Produce json
// @Param input body models.CreateLinkRequest true "Данные ссылки"
// @Success 200 {object} models.LinkResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/links [post]
func (h *LinkHandler) CreateShortLink(c *gin.Context) {
	var req models.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный URL"})
		return
	}

	userID := c.MustGet("userID").(int)

	var shortCode string
	var err error

	if req.CustomCode != "" {
		if !isValidCustomCode(req.CustomCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат кода"})
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
		UserID:      userID,
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
	}

	if err := h.LinkRepo.CreateLink(link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ссылки"})
		return
	}

	c.JSON(http.StatusOK, models.LinkResponse{
		ShortCode: shortCode,
		FullURL:   fmt.Sprintf("%s/%s", c.Request.Host, shortCode),
	})
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("short_code")
	log.Printf("[DEBUG] Запрос редиректа: %s", shortCode)

	link, err := h.LinkRepo.FindByShortCode(shortCode)
	if err != nil {
		log.Printf("[ERROR] Ошибка поиска: %v | Код: %s", err, shortCode)
		c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
		return
	}
	log.Printf("[INFO] Редирект: %s → %s", shortCode, link.OriginalURL)

	if err := h.LinkRepo.IncrementClickCount(shortCode); err != nil {
		log.Printf("[WARN] Ошибка инкремента: %v", err)
	}

	location, err := getLocationWithCache(c.ClientIP())
	if err != nil {
		log.Printf("[WARN] Ошибка геолокации: %v", err)
		location = "unknown"
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
	}

	if err := h.AnalyticRepo.SaveClick(clickData); err != nil {
		log.Printf("[ERROR] Ошибка сохранения клика: %v", err)
	}

	c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
}

// GetLinkStats godoc
// @Summary Получить статистику кликов
// @Description Возвращает аналитику кликов по короткой ссылке
// @Tags analytics
// @Security ApiKeyAuth
// @Produce json
// @Param short_code path string true "Короткий код ссылки" example(test123)
// @Success 200 {object} models.AnalyticsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/links/{short_code}/stats [get]
func (h *LinkHandler) GetLinkStats(c *gin.Context) {
	shortCode := c.Param("short_code")

	link, err := h.LinkRepo.FindByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
		return
	}

	dbStats, err := h.AnalyticRepo.GetAnalytics(link.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики"})
		return
	}

	response := models.AnalyticsResponse{
		TotalClicks: len(dbStats),
		Clicks:      make([]models.ClickStatistic, 0),
	}

	for _, s := range dbStats {
		response.Clicks = append(response.Clicks, models.ClickStatistic{
			IPAddress:  s.IPAddress,
			Location:   s.Location,
			DeviceType: s.DeviceType,
			OS:         s.OS,
			Browser:    s.Browser,
			ClickedAt:  s.ClickedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}
