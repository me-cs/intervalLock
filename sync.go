package IntervalLock

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type mutex struct {
	mu         *sync.Mutex
	doneC      *atomic.Int64
	forgetFunc func()
}

func locker(id string) *mutex {
	v, doneC, forgetFunc, _ := singleFlight(id, func() (any, error) {
		return &sync.Mutex{}, nil
	})
	return &mutex{
		mu:         v.(*sync.Mutex),
		doneC:      doneC,
		forgetFunc: forgetFunc,
	}
}

type call struct {
	wg         sync.WaitGroup
	val        interface{}
	err        error
	doneC      atomic.Int64
	forgetFunc func()
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type group struct {
	mu sync.Mutex // protects m
	m  sync.Map
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *group) Do(key string, fn func() (interface{}, error)) (interface{}, *atomic.Int64, func(), error) {
	if vCall, ok := g.m.Load(key); ok {
		c := vCall.(*call)
		c.doneC.Add(1)
		c.wg.Wait()
		return c.val, &c.doneC, c.forgetFunc, c.err
	}

	g.mu.Lock()
	if vCall, ok := g.m.Load(key); ok {
		g.mu.Unlock()
		c := vCall.(*call)
		c.doneC.Add(1)
		c.wg.Wait()
		return c.val, &c.doneC, c.forgetFunc, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m.Store(key, c)
	g.mu.Unlock()
	c.forgetFunc = sync.OnceFunc(func() {
		go func() {
			for {
				if c.doneC.Load() == 0 {
					g.m.Delete(key)
					return
				}
			}
		}()
		return
	})
	c.val, c.err = fn()
	c.doneC.Add(1)
	c.wg.Done()
	return c.val, &c.doneC, c.forgetFunc, c.err
}

var stdGroup = &group{}

func singleFlight(key string, fn func() (interface{}, error)) (interface{}, *atomic.Int64, func(), error) {
	return stdGroup.Do(key, fn)
}

func debug() {
	a := 0
	stdGroup.m.Range(func(key, value interface{}) bool {
		a++
		return true
	})
	fmt.Printf("Size=%v\n", a)
}
