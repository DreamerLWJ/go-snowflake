package snowflake

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	validTime := time.Now()
	invalidTime := time.Now().Add(time.Millisecond)

	var validWorkerID int64 = -1 ^ (-1 << 5)
	var invalidWorkerID int64 = -1 ^ (-1 << 5) + 1

	var validDataCenterID int64 = -1 ^ (-1 << 5)
	var invalidDataCenterID int64 = -1 ^ (-1 << 5) + 1

	g, err := NewStandardSnowflakeGenerator(validTime, validWorkerID, validDataCenterID)
	assert.Nil(t, err)
	assert.NotNil(t, g)

	g, err = NewStandardSnowflakeGenerator(invalidTime, validWorkerID, validDataCenterID)
	assert.Equal(t, ErrStartTimeInvalid, err)

	g, err = NewStandardSnowflakeGenerator(validTime, invalidWorkerID, validDataCenterID)
	assert.Equal(t, ErrWorkerIDOverLimit, err)

	g, err = NewStandardSnowflakeGenerator(validTime, validWorkerID, invalidDataCenterID)
	assert.Equal(t, ErrDataCenterIDOverLimit, err)

	c := Config{
		IDConfig: IDConfig{
			StartTime:        validTime,
			WorkerIDBits:     5,
			DataCenterIDBits: 5,
			SequenceIDBits:   12,
		},
		WorkerID:     validWorkerID,
		DataCenterID: validDataCenterID,
	}

	g, err = NewGenerator(c)
	assert.Nil(t, err)
	assert.NotNil(t, g)

	c.WorkerIDBits++
	g, err = NewGenerator(c)
	assert.Equal(t, ErrIDTooLong, err)
}

func TestGenerator_Next(t *testing.T) {
	var (
		startTime              = time.Now()
		workerID         int64 = 1
		dataCenterID     int64 = 1
		lastID           int64 = 0
		nextIDCollection       = make(map[int64]struct{}, 4000000)
	)

	g, err := NewStandardSnowflakeGenerator(startTime, workerID, dataCenterID)
	assert.Nil(t, err)

	for i := 0; i < 4000000; i++ {
		res, err := g.Next()
		assert.Nil(t, err)
		assert.Greater(t, res.NextID, lastID)

		assert.Equal(t, res.WorkerID, workerID)
		assert.Equal(t, res.DataCenterID, dataCenterID)

		_, exist := nextIDCollection[res.NextID]
		assert.False(t, exist)
		nextIDCollection[res.NextID] = struct{}{}
	}
}
