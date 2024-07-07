package db

import (
	"backend-go/models"
	"backend-go/utils"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log/slog"
	"os"
	"strconv"
)

var instance *gorm.DB = nil

var enableDebug = false

func EnableDebug() {
	enableDebug = true
}

func DisableDebug() {
	enableDebug = false
}

func GetDb() *gorm.DB {
	if instance != nil {
		if enableDebug {
			return instance.Debug()
		}
		return instance
	}
	var dbDialector gorm.Dialector = nil
	switch utils.GetConfigWithDefault("DB_TYPE", "mysql") {
	case "mysql":
		dbAddr := utils.GetConfigWithDefault("DB_SERVER", "localhost")
		dbPort, err := strconv.Atoi(utils.GetConfigWithDefault("DB_PORT", "3306"))
		if err != nil {
			slog.Error("Invalid database server port.")
			os.Exit(1)
		}
		dbUser := utils.GetConfigWithDefault("DB_USER", "root")
		dbPass := utils.GetConfigWithDefault("DB_PASSWORD", "root")
		dbName := utils.GetConfigWithDefault("DB_NAME", "novelcms")
		dbDialector = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbAddr, dbPort, dbName))
	case "sqlite":
		dbFile := utils.GetConfigWithDefault("DB_FILE", "data.db")
		dbDialector = sqlite.Open(dbFile)
	default:
		slog.Error("Invalid database type.")
		os.Exit(1)
	}
	var err error
	instance, err = gorm.Open(dbDialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: utils.GetConfigWithDefault("DB_TABLE_PREFIX", "novelcms_")},
	})
	if err != nil {
		slog.Error("Failed to open database.")
		os.Exit(1)
	}
	err = instance.AutoMigrate(&models.User{})
	if err != nil {
		slog.Error("Failed to migrate database structure.")
		os.Exit(1)
	}
	err = instance.AutoMigrate(&models.Article{})
	if err != nil {
		slog.Error("Failed to migrate database structure.")
		os.Exit(1)
	}
	err = instance.AutoMigrate(&models.Collection{})
	if err != nil {
		slog.Error("Failed to migrate database structure.")
		os.Exit(1)
	}
	if enableDebug {
		return instance.Debug()
	}
	return instance
}
