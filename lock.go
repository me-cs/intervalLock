package intervalLock

func Lock(key string) (unlock func()) {
	l := locker(key)
	l.mu.Lock()
	return func() {
		l.mu.Unlock()
		l.forgetFunc()
	}
}
