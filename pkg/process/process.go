package process

import "sync"

var done = make(chan bool)
var running = true
var m sync.Mutex

func Term() {
	m.Lock()
	defer m.Unlock()
	if !running {
		return
	}
	running = false
	close(done)
}

func Wait() {
	<-done
}
