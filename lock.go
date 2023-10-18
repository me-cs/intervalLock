package IntervalLock

func Lock(key string) (unlock func()) {
	l := locker(key)
	l.mu.Lock()
	return func() {
		l.mu.Unlock()
		l.doneC.Add(-1)
		l.forgetFunc()
	}
}
