package http

import "invest-robot/robot"

type HistoryHandler interface {
}

type DefaultHistoryHandler struct {
	h robot.HistoryAPI
}

func NewHistoryHandler(h robot.HistoryAPI) HistoryHandler {
	return DefaultHistoryHandler{h}
}
