package lazyLock

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
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
			t.Errorf("expect a=%v,got %v", conflict*(j+1), a)
		}
	}
	if a != times*conflict {
		t.Errorf("expect a=%v,got %v", times*conflict, a)
	}
	debugf()
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
			t.Errorf("expect a=%v,got %v", conflict*(j+1), a)
		}
	}
	if a != times*conflict {
		t.Errorf("expect a=%v,got %v", times*conflict, a)
	}
	debugf()
}

func TestMultiLock(t *testing.T) {
	var testLen = 1000
	nums := make([]int, testLen)
	var conflict = 1000 * testLen
	wg := sync.WaitGroup{}
	fmt.Printf("%d", conflict)
	for i := 0; i < conflict; i++ {
		wg.Add(1)
		go func(t int) {
			ind := t % testLen
			unlock := Lock(strconv.Itoa(ind))
			nums[ind]++
			unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
	for _, num := range nums {
		if num != conflict/testLen {
			t.Errorf("expect %v,got %v", conflict/testLen, num)
		}
	}
	debugf()
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

func TestDefer(t *testing.T) {
	x := make([]int, 1000)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				defer Lock(strconv.FormatInt(int64(j), 10))()
			}
			for j := 0; j < 1000; j++ {
				x[j]++
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 1000; i++ {
		if x[i] != 1000 {
			t.Errorf("expect %v,got %v", 1000, x[i])
		}
	}
}

func TestDeadLock(t *testing.T) {
	b := []int{2, 1}
	ch := make(chan bool)
	defer Lock("1")()
	go func() {
		for i := range b {
			defer Lock(strconv.Itoa(i))()
		}
		ch <- true
	}()
	af := time.After(1 * time.Second)
	select {
	case <-ch:
		t.Fatalf("it should be dead lock")
	case <-af:
	}
}
