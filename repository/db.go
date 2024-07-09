package repository

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var db *gorm.DB
var dbOnce sync.Once

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"database_name"`
}

func InitDatabase() *gorm.DB {
	dbOnce.Do(
		func() {
			dbConfig := DatabaseConfig{
				Host:         "localhost",
				Port:         3306,
				Username:     "dt",
				Password:     "dt",
				DatabaseName: "dt",
			}
			dbUsername := dbConfig.Username
			dsn := fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				dbUsername,
				dbConfig.Password,
				dbConfig.Host,
				dbConfig.Port,
				dbConfig.DatabaseName,
			)
			var err error
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
				//Logger: logger.Default.LogMode(logger.Info),
			})
			if err != nil {
				panic(fmt.Errorf("failed to connect database, error: %v", err))
			}

		},
	)

	return db
}

type TxI struct {
	DB  *gorm.DB
	Err error
}
