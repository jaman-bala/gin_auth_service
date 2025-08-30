package main

import (
	"gold_portal/config"
	"gold_portal/internal/api/route"
	"gold_portal/internal/infrastructure/database"

	_ "gold_portal/docs"

	"log"
)

// @title AUTH SERVICE API
// @version 1.0
// @description API for managing users and authentication in the Auth Service application.
// @host localhost:8080
func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	router := route.SetupRoutes(db, cfg)
	log.Printf("Сервер запущен на %p", router)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
