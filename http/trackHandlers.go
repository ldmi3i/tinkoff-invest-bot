package http

import "invest-robot/robot"

type TrackHandler interface {
}

type DefaultTrackHandler struct {
	t robot.TrackAPI
}

func NewTrackHandler(t robot.TrackAPI) TrackHandler {
	return DefaultTrackHandler{t}
}
