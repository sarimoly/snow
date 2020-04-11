package snow

import (
    "errors"
    "sync"
    "time"
)

type ModOption func(option *Option)

type Option struct {
    workIdBits     int
    dataCenterBits int
    sequenceBits   int
    twepoch        int64
}

var DefaultOption = Option{
    workIdBits:     5,
    dataCenterBits: 5,
    sequenceBits:   12,
    twepoch:        int64(1288834974657),
}

type Snow struct {
    workId       int64
    dataCenterId int64
    option       Option

    workerIDShift     int
    dataCenterIDShift int
    timestampShift    int

    seqMask uint32
    mu      sync.Mutex
    lastTs  int64
    seq     uint32
}

func NewSnow(workId, dataCenterId int64, modOption ...ModOption) (*Snow, error) {
    option := DefaultOption

    for _, fn := range modOption {
        fn(&option)
    }

    snow := Snow{
        workId:            workId,
        dataCenterId:      dataCenterId,
        option:            option,
        workerIDShift:     option.sequenceBits,
        dataCenterIDShift: option.sequenceBits + option.workIdBits,
        timestampShift:    option.sequenceBits + option.workIdBits + option.dataCenterBits,
        seqMask:           bitToMask(option.sequenceBits),
        lastTs:            -1,
        seq:               0,
    }

    if !snow.valid() {
        return nil, errors.New("bits invalid")
    }

    return &snow, nil
}

func (s *Snow) Gen(milliSec ...int64) (int64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    var now int64

    if len(milliSec) == 0 {
        now = time.Now().UnixNano() / 1e6
    } else {
        now = milliSec[0]
    }

    if now < s.lastTs {
        return -1, errors.New("time revert")
    }

    if now == s.lastTs {
        s.seq = (s.seq + 1) % bitToMask(s.option.sequenceBits)
        if s.seq == 0 {
            for now <= s.lastTs {
                now = time.Now().UnixNano() / 1e6
            }
        }
    } else {
        s.seq = 0
    }

    s.lastTs = now

    ts := now - s.option.twepoch
    ts <<= s.timestampShift

    sq := ts | int64(s.workId<<s.workerIDShift) | int64(s.dataCenterId<<s.dataCenterIDShift) | int64(s.seq)
    return sq, nil
}

func bitToMask(bit int) uint32 {
    return uint32(int32(-1) ^ (int32(-1) << bit))
}

func (s *Snow) valid() bool {
    if s.option.workIdBits+s.option.dataCenterBits+s.option.sequenceBits >= 64 {
        return false
    }

    if s.workId > int64(bitToMask(s.option.workIdBits)) {
        return false
    }

    if s.dataCenterId > int64(bitToMask(s.option.dataCenterBits)) {
        return false
    }

    return true
}
