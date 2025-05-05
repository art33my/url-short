package main

import (
	"database/sql"
	"fmt"
	"log"
	"url-short/internal/config"
	"url-short/internal/handlers"
	"url-short/internal/middleware"
	"url-short/internal/repositories"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	linkRepo := repositories.NewLinkRepository(db)
	analyticRepo := repositories.NewAnalyticRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	linkHandler := &handlers.LinkHandler{
		LinkRepo:     linkRepo,
		AnalyticRepo: analyticRepo,
	}

	r := gin.Default()

	r.Use(gin.Logger())

	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Connected to DB!"})
	})
	r.GET("/:short_code", linkHandler.Redirect)

	authGroup := r.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup.POST("/links", linkHandler.CreateShortLink)
	}

	authGroup.GET("/links/:short_code/stats", linkHandler.GetLinkStats)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
