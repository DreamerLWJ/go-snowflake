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
	sequence      int64 // 毫秒内序列
	lastTimestamp int64 // 上次生成 ID 的时间戳
}

// NewGenerator 获取雪花 ID 生成器
func NewGenerator(config Config) (*Generator, error) {
	g := Generator{Config: config}
	err := g.Init()
	if err != nil {
		return &g, err
	}
	return &g, nil
}

// NewStandardSnowflakeGenerator 获取标准的雪花 ID 生成器
func NewStandardSnowflakeGenerator(startTime time.Time, worker, dataCenterID int64) (*Generator, error) {
	g := Generator{Config: DefaultConfig(startTime, worker, dataCenterID)}
	err := g.Init()
	if err != nil {
		return &g, err
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
	if s.timestampLeftShift+41 > 63 {
		return ErrIDTooLong
	}
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
		s.lastTimestamp = timestamp
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
