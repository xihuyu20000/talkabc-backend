package utils

import (
	"strconv"
	"sync"
	"time"
)

const (
	workerIDBits     = 5
	dataCenterIDBits = 5
	sequenceBits     = 12

	maxWorkerID     = -1 ^ (-1 << workerIDBits)
	maxDataCenterID = -1 ^ (-1 << dataCenterIDBits)
	maxSequence     = -1 ^ (-1 << sequenceBits)

	workerIDShift     = sequenceBits
	dataCenterIDShift = sequenceBits + workerIDBits
	timestampShift    = sequenceBits + workerIDBits + dataCenterIDBits

	epoch = 1704067200000
)

type Snowflake struct {
	mu           sync.Mutex
	lastTimestamp int64
	workerID     int64
	dataCenterID int64
	sequence     int64
}

var snowflake *Snowflake

func init() {
	snowflake = NewSnowflake(1, 1)
}

func NewSnowflake(workerID, dataCenterID int64) *Snowflake {
	if workerID < 0 || workerID > maxWorkerID {
		panic("worker ID out of range")
	}
	if dataCenterID < 0 || dataCenterID > maxDataCenterID {
		panic("data center ID out of range")
	}
	return &Snowflake{
		lastTimestamp: 0,
		workerID:      workerID,
		dataCenterID:  dataCenterID,
		sequence:      0,
	}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.currentTimestamp()

	if timestamp < s.lastTimestamp {
		panic("clock moved backwards")
	}

	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			timestamp = s.waitNextMillis(timestamp)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = timestamp

	return (timestamp-epoch)<<timestampShift |
		s.dataCenterID<<dataCenterIDShift |
		s.workerID<<workerIDShift |
		s.sequence
}

func (s *Snowflake) currentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func (s *Snowflake) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := s.currentTimestamp()
	for timestamp <= lastTimestamp {
		timestamp = s.currentTimestamp()
	}
	return timestamp
}

func GenerateUID() string {
	return strconv.FormatInt(snowflake.NextID(), 10)
}