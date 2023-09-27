package main

import (
	"fmt"
	"strconv"
	"time"
)

type IDConfig struct {
	StartStamp       int64 // 开始时间戳
	WorkerIDBits     int64 // 机器 id 占用的长度
	DataCenterIDBits int64 // 机房 id 占用的长度
	SequenceIDBits   int64 // 序列占用的长度，1 毫秒内生成的 id 序列号
}

type Config struct {
	IDConfig

	WorkerID     int64 // 机器 ID
	DataCenterID int64 // 机房 ID
}

func DefaultConfig(worker, dataCenterID int64) Config {
	return Config{
		IDConfig: IDConfig{
			StartStamp:       0,
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

	workerIDShift      int64 // 机器 id 偏移量
	dataCenterIDShift  int64 // 机房 id 偏移量
	timestampLeftShift int64 // 时间戳偏移量
	maxWorkerID        int64 // 最大的机器 id
	maxDataCenterID    int64 // 最大机房 id
	sequenceMask       int64 // 生成序列的掩码(用于限制)

	sequence      int64 // 毫秒内序列 TODO 可见性问题
	lastTimestamp int64 // 上次生成 ID 的时间戳 TODO 可见性问题
}

// NewStandardSnowflakeGenerator 获取标准的雪花 ID 生成器
func NewStandardSnowflakeGenerator(c Config) (*Generator, error) {
	g := Generator{Config: c}
	g.Init()
	if g.timestampLeftShift > 63 {
		return &g, ErrIDTooLong
	}

	return &g, nil
}

func (s *Generator) Init() {
	s.maxWorkerID = -1 ^ (-1 << s.WorkerIDBits)
	s.maxDataCenterID = -1 ^ (-1 << s.DataCenterIDBits)
	s.workerIDShift = s.SequenceIDBits
	s.dataCenterIDShift = s.SequenceIDBits + s.WorkerIDBits
	s.timestampLeftShift = 41 + s.WorkerIDBits + s.DataCenterIDBits
	s.sequenceMask = -1 ^ (-1 << s.SequenceIDBits)
	s.sequence = 0
	s.lastTimestamp = -1
}

func (s *Generator) Next() (res ID, err error) {
	timestamp := time.Now().Unix()

	if timestamp < s.lastTimestamp {
		return res, ErrClockMovedBack
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & s.sequenceMask
		if s.sequence == 0 {
			timestamp = s.tilNextMills(s.lastTimestamp)
		}
	} else {
		s.sequence = 0
	}

	nextID := ((timestamp - s.StartStamp) << s.timestampLeftShift) | (s.DataCenterID << s.dataCenterIDShift) | (s.WorkerID << s.workerIDShift) | s.sequence
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

func main() {
	workerIDBits := 5
	fmt.Println(intToBitString(workerIDBits))
	fmt.Println(intToBitString(-1))
	fmt.Println(intToBitString(-1 << workerIDBits))
	fmt.Println(-1 ^ (-1 << workerIDBits))
}

func intToBitString(num int) string {
	// 获取整数的位数
	bitCount := strconv.IntSize

	// 创建一个存储位的切片
	bits := make([]byte, bitCount)

	// 使用循环逐位填充切片
	for i := 0; i < bitCount; i++ {
		if num&(1<<uint(i)) != 0 {
			bits[bitCount-i-1] = '1'
		} else {
			bits[bitCount-i-1] = '0'
		}
	}

	// 将切片转换为字符串
	return string(bits)
}
