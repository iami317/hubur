package hubur

import (
	"context"
	"math"
	"sync"
)

type SizedWaitGroup struct {
	wg      sync.WaitGroup
	current chan struct{}
}

func NewSizedWaitGroup(limit int) SizedWaitGroup {
	size := math.MaxInt32 // 2^32 - 1
	if limit > 0 {
		size = limit
	}
	if limit <= 0 {
		panic("limit must >0")
	}
	return SizedWaitGroup{
		wg:      sync.WaitGroup{},
		current: make(chan struct{}, size),
	}
}

func (s *SizedWaitGroup) Add() {
	s.current <- struct{}{}
	s.wg.Add(1)
}

func (s *SizedWaitGroup) AddWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.current <- struct{}{}:
		break
	}
	s.wg.Add(1)
	return nil
}

func (s *SizedWaitGroup) Done() {
	<-s.current
	s.wg.Done()
}

func (s *SizedWaitGroup) Wait() {
	s.wg.Wait()
}
