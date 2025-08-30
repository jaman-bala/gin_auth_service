package database

import (
	"fmt"
	"gold_portal/config"
	"gold_portal/internal/domain/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid-ossp extension: %v", err)
	}

	err = db.AutoMigrate(
		&entities.User{},
		&entities.AuditLog{},
	)
	if err != nil {
		return nil, err
	}

	createDefaultAdmin(db)
	return db, nil
}

func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&entities.User{}).Where("phone = ?", "+996500500500").Count(&count)
	if count > 0 {
		return
	}

	adminUser := entities.User{
		Password: "Password123",
		Phone:    "+996500500500",
		Role:     entities.RoleSuperUser,
		IsActive: true,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		fmt.Printf("Ошибка создания администратора: %v\n", err)
	}
}
