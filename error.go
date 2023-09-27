package main

import "errors"

var (
	ErrClockMovedBack = errors.New("machine clock moved back")
	ErrIDTooLong      = errors.New("config result id too long, id should shorter than 64 bits")
)
