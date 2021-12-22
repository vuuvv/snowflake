package snowflake

import (
	"github.com/vuuvv/errors"
	"sync"
	"time"
)

type ID struct {
	value int64
}

type Snowflake struct {
	mutex          sync.Mutex
	workerId       int64 // 机器标识
	epoch          int64 // 开始时间戳
	sequenceBits   int64 // 序列号占用位数
	workerIdBits   int64 // 机器标识占用位数
	maxWorkers     int64 // 机器最大值
	maxSequence    int64 // 每毫秒序列号最大值
	workerShift    int64 // worker向左位移数
	timestampShift int64 // sequence向左位移数
	sequence       int64 // 当前毫秒的序列号
	lastStamp      int64 // 时间回拨处理：上一次时间戳
	stepSize       int64 // 时间回拨处理：时间回拨后sequence的位置调整
	baseSequence   int64 // 时间回拨处理：基础序列号, 每发生一次时钟回拨即改变, basicSequence += stepSize
}

type Option func(snowflake *Snowflake)

func WithEpoch(epoch time.Time) Option {
	return func(snowflake *Snowflake) {
		snowflake.epoch = epoch.UnixMilli()
	}
}

func WithSequenceBits(sequenceBits int64) Option {
	return func(snowflake *Snowflake) {
		snowflake.sequenceBits = sequenceBits
	}
}

func WithWorkerIdBits(workerIdBits int64) Option {
	return func(snowflake *Snowflake) {
		snowflake.workerIdBits = workerIdBits
	}
}

func WithWorkerId(workerId int64) Option {
	return func(snowflake *Snowflake) {
		snowflake.workerId = workerId
	}
}

func NewSnowflake(opts ...Option) (s *Snowflake, err error) {
	time.Now()
	s = &Snowflake{
		epoch:        time.Date(2021, 12, 21, 0, 0, 0, 0, time.UTC).UnixMilli(),
		workerIdBits: 8,
		sequenceBits: 12,
	}

	for _, o := range opts {
		o(s)
	}

	s.maxWorkers = -1 ^ (-1 << s.workerIdBits)
	s.maxSequence = -1 ^ (-1 << s.sequenceBits)
	s.workerShift = s.sequenceBits
	s.timestampShift = s.sequenceBits + s.workerIdBits
	s.sequence = 0
	s.lastStamp = -1
	s.stepSize = s.maxSequence >> 1
	s.baseSequence = 0

	if s.workerId > s.maxWorkers {
		return nil, errors.New("datacenterId can't be greater than MAX_DATACENTER_NUM or less than 0")
	}

	return
}

func (s *Snowflake) SetWorkerId(workerId int64) {
	s.workerId = workerId
}

func (s *Snowflake) GetMaxWorks() int64 {
	return s.maxWorkers
}

func (s *Snowflake) Next() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stamp := currentMilli()
	if stamp < s.lastStamp {
		return s.handleClockBackwards(stamp)
	}

	if stamp == s.lastStamp {
		// 相同毫秒内，序列号自增，最大值为stepSize
		s.sequence = (s.sequence + 1) & s.stepSize
		if s.sequence == 0 {
			stamp = s.nextMilli()
		}
	} else {
		s.sequence = s.baseSequence
	}

	s.lastStamp = stamp

	return (stamp-s.epoch)<<s.timestampShift |
		s.workerId<<s.workerShift |
		s.sequence
}

func (s *Snowflake) handleClockBackwards(stamp int64) int64 {
	s.baseSequence += s.stepSize
	if s.baseSequence == s.maxSequence+1 {
		s.baseSequence = 0
		stamp = s.nextMilli()
	}
	s.sequence = s.baseSequence
	s.lastStamp = stamp
	return (stamp-s.epoch)<<s.timestampShift |
		s.workerId<<s.workerShift |
		s.sequence
}

// 出现时间回拨的现象，
func (s *Snowflake) nextMilli() int64 {
	milli := currentMilli()
	for milli <= s.lastStamp {
		time.Sleep(time.Millisecond*time.Duration(s.lastStamp-milli) + 1)
		milli = currentMilli()
	}
	return milli
}

func currentMilli() int64 {
	return time.Now().UnixMilli()
}
