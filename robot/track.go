package robot

type TrackAPI interface {
}

type DefaultTrackAPI struct {
}

func NewTrackAPI() TrackAPI {
	return DefaultHistoryAPI{}
}
