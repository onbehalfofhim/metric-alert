package service

import "errors"

var (
	ErrInvalidType  = errors.New("unknown metric type")
	ErrInvalidValue = errors.New("unknown metric value")
	ErrInvalidName  = errors.New("invalid metric name (empty name)")
)
