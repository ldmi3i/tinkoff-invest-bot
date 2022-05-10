package helper

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"invest-robot/domain"
	"log"
)

var db *gorm.DB

func initDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		GetDbHost(), GetDbUser(), GetDbPasswd(), GetDbName(), GetDbPort())
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
