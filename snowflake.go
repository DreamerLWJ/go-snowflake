package snowflake

import (
	"sync"
	"time"
)

type IDConfig struct {
	StartTime        time.Time // 开始时间
	WorkerIDBits     int64     // 机器 id 占用的长度
	DataCenterIDBits int64     // 机房 id 占用的长度
	SequenceIDBits   int64     // 序列占用的长度，1 毫秒内生成的 id 序列号
}

type Config struct {
	IDConfig

	WorkerID     int64 // 机器 ID
	DataCenterID int64 // 机房 ID
}

func DefaultConfig(startTime time.Time, worker, dataCenterID int64) Config {
	return Config{
		IDConfig: IDConfig{
			StartTime:        startTime,
			WorkerIDBits:     5,
			DataCenterIDBits: 5,
			SequenceIDBits:   12,
		},
		WorkerID:     worker,
		DataCenterID: dataCenterID,
	}
}

type ID struct {
	NextID       int64 // 雪花 ID
	Timestamp    int64 // 生成的时间戳
	WorkerID     int64 // 机器 ID
	DataCenterID int64 // 机房 ID
	Sequence     int64 // 序列号
}

type Generator struct {
	Config

	startTimeStamp     int64 // 开始的时间戳
	workerIDShift      int64 // 机器 id 偏移量
	dataCenterIDShift  int64 // 机房 id 偏移量
	timestampLeftShift int64 // 时间戳偏移量
	maxWorkerID        int64 // 最大的机器 id
	maxDataCenterID    int64 // 最大机房 id
	sequenceMask       int64 // 生成序列的掩码(用于限制)

	mu            sync.Mutex
	sequence      int64 // 毫秒内序列 TODO 可见性问题
	lastTimestamp int64 // 上次生成 ID 的时间戳 TODO 可见性问题
}

// NewStandardSnowflakeGenerator 获取标准的雪花 ID 生成器
func NewStandardSnowflakeGenerator(c Config) (*Generator, error) {
	g := Generator{Config: c}
	err := g.Init()
	if err != nil {
		return &g, err
	}
	if g.timestampLeftShift > 63 {
		return &g, ErrIDTooLong
	}

	return &g, nil
}

func (s *Generator) Init() error {
	if s.StartTime.After(time.Now()) {
		return ErrStartTimeInvalid
	}
	s.startTimeStamp = s.StartTime.UnixMilli()
	s.maxWorkerID = -1 ^ (-1 << s.WorkerIDBits)
	s.maxDataCenterID = -1 ^ (-1 << s.DataCenterIDBits)
	s.workerIDShift = s.SequenceIDBits
	s.dataCenterIDShift = s.SequenceIDBits + s.WorkerIDBits
	s.timestampLeftShift = s.SequenceIDBits + s.WorkerIDBits + s.DataCenterIDBits
	s.sequenceMask = -1 ^ (-1 << s.SequenceIDBits)
	s.sequence = 0
	s.lastTimestamp = -1

	if s.Config.WorkerID > s.maxWorkerID {
		return ErrWorkerIDOverLimit
	}
	if s.Config.DataCenterID > s.maxDataCenterID {
		return ErrDataCenterIDOverLimit
	}
	return nil
}

func (s *Generator) Next() (res ID, err error) {
	timestamp := time.Now().UnixMilli()

	s.mu.Lock()
	if timestamp < s.lastTimestamp {
		return res, ErrClockMovedBack
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & s.sequenceMask
		if s.sequence == 0 {
			s.lastTimestamp = s.tilNextMills(s.lastTimestamp)
			timestamp = s.lastTimestamp
		}
	} else {
		s.sequence = 0
	}
	s.mu.Unlock()

	nextID := ((timestamp - s.startTimeStamp) << s.timestampLeftShift) | (s.DataCenterID << s.dataCenterIDShift) | (s.WorkerID << s.workerIDShift) | s.sequence
	return ID{
		NextID:       nextID,
		Timestamp:    timestamp,
		WorkerID:     s.WorkerID,
		DataCenterID: s.DataCenterID,
		Sequence:     s.sequence,
	}, nil
}

func (s *Generator) tilNextMills(lastTimestamp int64) (nextTimestamp int64) {
	ts := time.Now().Unix()
	for ts <= lastTimestamp {
		ts = time.Now().Unix()
	}
	return ts
}
