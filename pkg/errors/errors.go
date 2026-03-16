package errors

import "errors"

var (
	ErrMetricNotFound = errors.New("metric not found")
	ErrInvalidType    = errors.New("unknown metric type")
	ErrInvalidValue   = errors.New("unknown metric value")
)
