package strategy

import "errors"

type Type int

const (
	Median Type = iota
)

func Load(id uint) (*Strategy, error) {
	return nil, errors.New("not implemented")
}
