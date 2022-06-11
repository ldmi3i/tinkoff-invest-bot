package app

import (
	"invest-robot/internal/helper"
	"invest-robot/pkg/robot"
	"invest-robot/pkg/web"
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

	robot.Init()
	robot.StartBgTasks()
	web.StartHttp()
}
