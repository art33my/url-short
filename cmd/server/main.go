package main

import (
	"database/sql"
	"fmt"
	"log"
	"url-short/internal/config"
	"url-short/internal/handlers"
	"url-short/internal/middleware"
	"url-short/internal/repositories"

	_ "url-short/docs"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title URL Shortener API
// @version 1.0
// @description API для сокращения URL-адресов
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg := config.LoadConfig()

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBName == "" {
		log.Fatal("[FATAL] Не заданы параметры подключения к БД в .env")
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("[FATAL] Ошибка подключения: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("[FATAL] Ошибка аутентификации: %v", err)
	}
	log.Println("Успешное подключение к PostgreSQL")

	userRepo := repositories.NewUserRepository(db)
	linkRepo := repositories.NewLinkRepository(db)
	analyticRepo := repositories.NewAnalyticRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	linkHandler := &handlers.LinkHandler{
		LinkRepo:     linkRepo,
		AnalyticRepo: analyticRepo,
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	r.GET("/:short_code", linkHandler.Redirect)
	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
	}

	authGroup := api.Group("")
	authGroup.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup.POST("/links", linkHandler.CreateShortLink)
	}

	statsGroup := api.Group("")
	statsGroup.Use(middleware.AuthMiddleware(cfg))
	{
		statsGroup.GET("/links/:short_code/stats", linkHandler.GetLinkStats)
	}
	// swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("Сервер запущен на http://localhost:%s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("[FATAL] Ошибка запуска: %v", err)
	}
}
