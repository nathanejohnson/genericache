package genericache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type noErrorFilled struct {
	called int64
	delay  time.Duration
	t      *testing.T
}

func (nof *noErrorFilled) fill(i int) (int, error) {
	atomic.AddInt64(&nof.called, 1)
	if nof.delay > 0 {
		nof.t.Logf("%d sleeping %s", i, nof.delay.String())
		time.Sleep(nof.delay)
	}
	nof.t.Logf("returning from %d", i)
	return i + 5, nil
}

type errorUntilFilled struct {
	errUntilCount int64
	called        int64
	delay         time.Duration
	t             *testing.T
}

var ErrFill = errors.New("this is a fill error")

func (euf *errorUntilFilled) fill(i int) (int, error) {
	v := atomic.AddInt64(&euf.called, 1)
	if euf.delay > 0 {
		euf.t.Logf("%d sleeping %s", i, euf.delay.String())
		time.Sleep(euf.delay)
	}
	if v <= euf.errUntilCount {
		euf.t.Logf("%d error count not reached, at %d limit %d", i, v, euf.errUntilCount)
		return 0, ErrFill
	}
	euf.t.Logf("returning from %d", i)
	return i + 5, nil
}

func TestNoErrors(t *testing.T) {
	nof := &noErrorFilled{
		t:     t,
		delay: time.Second * 3,
	}
	c := NewGeneriCache(nof.fill, false)
	var wg sync.WaitGroup
	t1 := time.Now()
	for i := 1; i < 10; i++ {
		for j := 0; j < 3; j++ {
			wg.Add(1)
			go func(i int) {
				v, _ := c.Get(i)
				t.Logf("fetched %d and got %d", i, v)
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	t.Logf("time elapsed was %s, shoudl have been approximately 3 seconds", time.Now().Sub(t1).String())
}

func TestWithErrors(t *testing.T) {
	euf := &errorUntilFilled{
		errUntilCount: 12,
		delay:         time.Second * 3,
		t:             t,
	}
	c := NewGeneriCache(euf.fill, true)
	var wg sync.WaitGroup
	t1 := time.Now()
	for i := 1; i < 10; i++ {
		for j := 0; j < 4; j++ {
			wg.Add(1)
			go func(i int) {
				v, err := c.Get(i)
				t.Logf("fetched %d and got %d and error %#v", i, v, err)
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	t.Logf("time elapsed was %s, shoudl have been approximately 9 seconds", time.Now().Sub(t1).String())

}
