package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"invest-robot/internal/domain"
	"invest-robot/internal/env"
	"log"
)

var db *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		env.GetDbHost(), env.GetDbUser(), env.GetDbPasswd(), env.GetDbName(), env.GetDbPort())
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: 1000,
	})
	if err != nil {
		log.Fatalf("Error while connecting to db: \n%s", err)
	}
	db = dbConn
	err = migrate()
	if err != nil {
		log.Fatalf("Error while performing migrations:\n%s", err)
	}
}

func migrate() error {
	return db.AutoMigrate(
		&domain.History{},
		&domain.Algorithm{},
		&domain.Action{},
		&domain.Param{},
		&domain.CtxParam{},
		&domain.MoneyLimit{},
	)
}

func GetDB() *gorm.DB {
	return db
}