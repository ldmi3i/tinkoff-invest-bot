package main

import (
	"github.com/ldmi3i/tinkoff-invest-bot/bot"
	"github.com/ldmi3i/tinkoff-invest-bot/web"
)

func main() {
	defer func() {
		bot.PostProcess()
	}()

	bot.Init()
	bot.StartBgTasks()
	web.StartHttp()
}
