package app

import (
	"invest-robot/helper"
	"invest-robot/http"
	"invest-robot/robot"
	"log"
)

func Start() {
	defer func() {
		robot.PostProcess()
		err := helper.Close()
		if err != nil {
			log.Printf("error while closing grpc connection: \n%s", err)
		}
	}()

	robot.StartBgTasks()
	http.StartHttp()
}
