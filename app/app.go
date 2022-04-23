package app

import (
	"invest-robot/helper"
	"invest-robot/service"
	"invest-robot/tinapi"
	"log"
)

func Start() {
	defer func() {
		err := helper.Close()
		if err != nil {
			log.Printf("error while closing grpc connection: \n%s", err)
		}
	}()

	api := tinapi.NewTinApi()
	infoSrv := service.NewInfoSrv(api)
	book, err := infoSrv.GetOrderBook()
	if err != nil {
		log.Printf("Error getting orders:\n%s", err)
		return
	}
	log.Printf("Successfully retrieved book: %s", book)
}
