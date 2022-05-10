package app

import (
	"invest-robot/helper"
	"invest-robot/http"
	"invest-robot/robot"
	"log"
)

func Start() {
	defer func() {
		err := helper.Close()
		if err != nil {
			log.Printf("error while closing grpc connection: \n%s", err)
		}
	}()

	//api := tinapi.NewTinApi()
	//infoSrv := service.NewInfoService(api)
	//book, err := infoSrv.GetOrderBook()
	//if err != nil {
	//	log.Printf("Error getting orders:\n%s", err)
	//	return
	//}
	//log.Printf("Successfully retrieved book: %s", book)

	robot.StartBgTasks()
	http.StartHttp()
}
