package helper

import (
	"log"
	"os"
)

var tinToken string
var grpcAddr string
var dbUser string
var dbPasswd string
var dbHost string
var dbPort string
var dbName string

var srvHost string
var srvPort string

func initEnv() {
	sanityCheck("TIN_TOKEN")
	tinToken = os.Getenv("TIN_TOKEN")

	grpcAddr = getOrDefault("TIN_ADDRESS", "invest-public-api.tinkoff.ru:443")

	dbUser = getOrDefault("DB_USER", "postgres")
	dbPasswd = getOrDefault("DB_PASSWD", "postgres")
	dbHost = getOrDefault("DB_HOST", "localhost")
	dbPort = getOrDefault("DB_PORT", "5432")
	dbName = getOrDefault("DB_NAME", "invest-bot")

	srvHost = getOrDefault("SERVER_HOST", "localhost")
	srvPort = getOrDefault("SERVER_PORT", "8017")
}

func getOrDefault(env string, def string) string {
	if res, ok := os.LookupEnv(env); ok {
		return res
	}
	return def
}

func sanityCheck(params ...string) {
	for _, param := range params {
		if param == "" {
			log.Fatalf("Environment variable '%s' not defined", param)
		}
	}
}

func GetTinToken() string {
	return tinToken
}

func GetGRPCAddress() string {
	return grpcAddr
}

func GetDbUser() string {
	return dbUser
}

func GetDbPasswd() string {
	return dbPasswd
}

func GetDbHost() string {
	return dbHost
}

func GetDbPort() string {
	return dbPort
}

func GetDbName() string {
	return dbName
}

func GetSrvHost() string {
	return srvHost
}

func GetSrvPort() string {
	return srvPort
}
