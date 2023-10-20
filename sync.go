package lazyLock

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
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
	m  map[string]*call
}

// errGoexit indicates the runtime.Goexit was called in
// the user given function.
var errGoexit = errors.New("runtime.Goexit was called")

// A panicError is an arbitrary value recovered from a panic
// with the stack trace during the execution of given function.
type panicError struct {
	value interface{}
	stack []byte
}

// Error implements error interface.
func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.value, p.stack)
}

func (p *panicError) Unwrap() error {
	err, ok := p.value.(error)
	if !ok {
		return nil
	}

	return err
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.

func (g *group) Do(key string, fn func() (interface{}, error)) (v interface{}, at *atomic.Int64, ff func(), err error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.doneC.Add(1)
		g.mu.Unlock()
		c.wg.Wait()

		if e, ok := c.err.(*panicError); ok {
			panic(e)
		} else if c.err == errGoexit {
			runtime.Goexit()
		}
		return c.val, &c.doneC, c.forgetFunc, c.err
	}
	c := new(call)
	c.wg.Add(1)
	c.doneC.Add(1)
	c.forgetFunc = func() {
		g.mu.Lock()
		if c.doneC.Add(-1) == 0 {
			delete(g.m, key)
		}
		g.mu.Unlock()
	}
	g.m[key] = c
	g.mu.Unlock()
	g.doCall(c, key, fn)
	return c.val, &c.doneC, c.forgetFunc, c.err
}

// doCall handles the single call for a key.
func (g *group) doCall(c *call, key string, fn func() (interface{}, error)) {
	normalReturn := false
	recovered := false

	// use double-defer to distinguish panic from runtime.Goexit,
	// more details see https://golang.org/cl/134395
	defer func() {
		// the given function invoked runtime.Goexit
		if !normalReturn && !recovered {
			c.err = errGoexit
		}

		g.mu.Lock()
		defer g.mu.Unlock()
		c.wg.Done()
		//if g.m[key] == c {
		//	delete(g.m, key)
		//}

		if e, ok := c.err.(*panicError); ok {
			// In order to prevent the waiting channels from being blocked forever,
			// needs to ensure that this panic cannot be recovered.
			panic(e)
		} else if c.err == errGoexit {
			// Already in the process of goexit, no need to call again
		}
	}()

	func() {
		defer func() {
			if !normalReturn {
				// Ideally, we would wait to take a stack trace until we've determined
				// whether this is a panic or a runtime.Goexit.
				//
				// Unfortunately, the only way we can distinguish the two is to see
				// whether the recover stopped the goroutine from terminating, and by
				// the time we know that, the part of the stack trace relevant to the
				// panic has been discarded.
				if r := recover(); r != nil {
					c.err = newPanicError(r)
				}
			}
		}()

		c.val, c.err = fn()
		normalReturn = true
	}()

	if !normalReturn {
		recovered = true
	}
}

func newPanicError(v interface{}) error {
	stack := debug.Stack()

	// The first line of the stack trace is of the form "goroutine N [status]:"
	// but by the time the panic reaches Do the goroutine may no longer exist
	// and its status will have changed. Trim out the misleading line.
	if line := bytes.IndexByte(stack[:], '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{value: v, stack: stack}
}

var stdGroup = &group{}

func singleFlight(key string, fn func() (interface{}, error)) (interface{}, *atomic.Int64, func(), error) {
	return stdGroup.Do(key, fn)
}

func debugf() {
	fmt.Printf("Size=%v\n", len(stdGroup.m))
}
