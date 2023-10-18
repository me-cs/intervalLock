package intervalLock

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	//"golang.org/x/sync/singleflight"
)

type mutex struct {
	mu         *sync.Mutex
	doneC      *atomic.Int64
	forgetFunc func()
}

func locker(id string) *mutex {
	v, doneC, forgetFunc, _, _ := singleFlight(id, func() (any, error) {
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
	dups       int
	chans      []chan<- Result
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// Result holds the results of Do, so they can be passed
// on a channel.
type Result struct {
	Val    interface{}
	Err    error
	Shared bool
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

var counter = int64(0)
var counter1 = int64(0)

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.

func (g *group) Do(key string, fn func() (interface{}, error)) (v interface{}, at *atomic.Int64, ff func(), err error, shared bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		c.doneC.Add(1)
		g.mu.Unlock()
		c.wg.Wait()

		if e, ok := c.err.(*panicError); ok {
			panic(e)
		} else if c.err == errGoexit {
			runtime.Goexit()
		}
		return c.val, &c.doneC, c.forgetFunc, c.err, true
	}
	c := new(call)
	c.wg.Add(1)
	c.doneC.Add(1)
	c.forgetFunc = func() {
		g.mu.Lock()
		if c.doneC.Add(-1) == 0 {
			atomic.AddInt64(&counter1, 1)
			delete(g.m, key)
		}
		g.mu.Unlock()
	}
	//c.forgetFunc = sync.OnceFunc(func() {
	//	go func() {
	//		t := time.NewTicker(time.Second)
	//		defer t.Stop()
	//		for range t.C {
	//			g.mu.Lock()
	//			if c.doneC.Load() == 0 {
	//				delete(g.m, key)
	//				g.mu.Unlock()
	//				atomic.AddInt64(&counter1, 1)
	//				return
	//			}
	//			g.mu.Unlock()
	//		}
	//	}()
	//	return
	//})
	g.m[key] = c
	g.mu.Unlock()
	atomic.AddInt64(&counter, 1)
	g.doCall(c, key, fn)
	return c.val, &c.doneC, c.forgetFunc, c.err, c.dups > 0
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
			if len(c.chans) > 0 {
				go panic(e)
				select {} // Keep this goroutine around so that it will appear in the crash dump.
			} else {
				panic(e)
			}
		} else if c.err == errGoexit {
			// Already in the process of goexit, no need to call again
		} else {
			// Normal return
			for _, ch := range c.chans {
				ch <- Result{c.val, c.err, c.dups > 0}
			}
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

//func (g *group) Do(key string, fn func() (interface{}, error)) (interface{}, *atomic.Int64, func(), error) {
//	if vCall, ok := g.m.Load(key); ok {
//		c := vCall.(*call)
//		c.doneC.Add(1)
//		c.wg.Wait()
//		return c.val, &c.doneC, c.forgetFunc, c.err
//	}
//
//	g.mu.Lock()
//	if vCall, ok := g.m.Load(key); ok {
//		g.mu.Unlock()
//		c := vCall.(*call)
//		c.doneC.Add(1)
//		c.wg.Wait()
//		return c.val, &c.doneC, c.forgetFunc, c.err
//	}
//
//	c := new(call)
//	c.wg.Add(1)
//	g.m.Store(key, c)
//	g.mu.Unlock()
//	c.forgetFunc = sync.OnceFunc(func() {
//		go func() {
//			for {
//				if c.doneC.Load() == 0 {
//					g.m.Delete(key)
//					atomic.AddInt64(&counter1, 1)
//					return
//				}
//			}
//		}()
//		return
//	})
//	atomic.AddInt64(&counter, 1)
//	c.val, c.err = fn()
//	c.doneC.Add(1)
//	c.wg.Done()
//	return c.val, &c.doneC, c.forgetFunc, c.err
//}

var stdGroup = &group{}

func singleFlight(key string, fn func() (interface{}, error)) (interface{}, *atomic.Int64, func(), error, bool) {
	return stdGroup.Do(key, fn)
}

func debugf() {

	fmt.Printf("Size=%v\n", len(stdGroup.m))
}

// Forget tells the singleflight to forget about a key.  Future calls
// to Do for this key will call the function rather than waiting for
// an earlier call to complete.
func (g *group) Forget(key string) {
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}
