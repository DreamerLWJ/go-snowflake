package snowflake

import "errors"

var (
	ErrClockMovedBack        = errors.New("machine clock moved back")
	ErrIDTooLong             = errors.New("config result id too long, id should shorter than 64 bits")
	ErrWorkerIDOverLimit     = errors.New("workerID exceeds the configured workerIDBits size")
	ErrDataCenterIDOverLimit = errors.New("dataCenterID exceeds the configured DataCenterIDBits size")
	ErrStartTimeInvalid      = errors.New("start time invalid")
)
