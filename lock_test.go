package intervalLock

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	var a = 0
	var times = 10
	var conflict = 100000
	for j := 0; j < times; j++ {
		time.Sleep(time.Millisecond * 50)
		wg := sync.WaitGroup{}
		for i := 0; i < conflict; i++ {
			wg.Add(1)
			go func() {
				unlock := Lock("1")
				a++
				unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		if a != conflict*(j+1) {
			t.Errorf("expect a=%v,got %v", a, conflict*(j+1))
		}
	}
	if a != times*conflict {
		t.Errorf("expect a=%v,got %v", a, times*conflict)
	}
	time.Sleep(time.Millisecond * 100)
	debugf()
	fmt.Printf("used %d mutex\n", atomic.LoadInt64(&counter))
	fmt.Printf("closed %d mutex\n", atomic.LoadInt64(&counter1))
}

func TestLargeInterval(t *testing.T) {
	var a = 0
	var times = 100
	var conflict = 10
	for j := 0; j < times; j++ {
		time.Sleep(time.Millisecond * 10)
		wg := sync.WaitGroup{}
		for i := 0; i < conflict; i++ {
			wg.Add(1)
			go func() {
				unlock := Lock("1")
				a++
				unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		if a != conflict*(j+1) {
			t.Errorf("expect a=%v,got %v", a, conflict*(j+1))
		}
	}
	if a != times*conflict {
		t.Errorf("expect a=%v,got %v", a, times*conflict)
	}
	time.Sleep(time.Millisecond * 100)
	debugf()
	fmt.Printf("used %d mutex", atomic.LoadInt64(&counter))
	fmt.Printf("closed %d mutex", atomic.LoadInt64(&counter1))
}

func benchmarkMutex(b *testing.B, slack, work bool) {
	if slack {
		b.SetParallelism(10)
	}
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			Lock("test")()
			if work {
				for i := 0; i < 100; i++ {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}

func BenchmarkMutex(b *testing.B) {
	benchmarkMutex(b, false, false)
}

func HammerMutex(loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		Lock("test")()
	}
	cdone <- true
}

func TestMutex(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)

	Lock("test")()
	Lock("test")()

	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}

func TestMutexFairness(t *testing.T) {
	stop := make(chan bool)
	defer close(stop)
	go func() {
		for {
			unlock := Lock("test")
			time.Sleep(100 * time.Microsecond)
			unlock()
			select {
			case <-stop:
				return
			default:
			}
		}
	}()
	done := make(chan bool, 1)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Microsecond)
			Lock("test")()
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("can't acquire Mutex in 10 seconds")
	}
}

func BenchmarkMutexSpin(b *testing.B) {
	// This benchmark models a situation where spinning in the mutex should be
	// profitable. To achieve this we create a goroutine per-proc.
	// These goroutines access considerable amount of local data so that
	// unnecessary rescheduling is penalized by cache misses.
	var acc0, acc1 uint64
	b.RunParallel(func(pb *testing.PB) {
		var data [16 << 10]uint64
		for i := 0; pb.Next(); i++ {
			unlock := Lock("test")
			acc0 -= 100
			acc1 += 100
			unlock()
			for i := 0; i < len(data); i += 4 {
				data[i]++
			}
		}
	})
}
