package intervalLock

import (
	"sync"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	var a = 0
	var times = 10
	var conflict = 1000
	for j := 0; j < times; j++ {
		time.Sleep(time.Millisecond * 50)
		wg := sync.WaitGroup{}
		for i := 0; i < conflict; i++ {
			wg.Add(1)
			func() {
				defer wg.Done()
				unlock := Lock("1")
				a++
				unlock()
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
	debug()
}

func TestLargeInterval(t *testing.T) {
	var a = 0
	var times = 20
	var conflict = 100
	for j := 0; j < times; j++ {
		time.Sleep(time.Millisecond * 50)
		wg := sync.WaitGroup{}
		for i := 0; i < conflict; i++ {
			wg.Add(1)
			func() {
				defer wg.Done()
				unlock := Lock("1")
				a++
				unlock()
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
	debug()
}
