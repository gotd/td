package pool

import "sync"

type dcID int

type dcInfo struct {
	primaryDC   dcID
	primaryAddr string
	primaryMux  sync.RWMutex
}

func (d *dcInfo) Load() (id dcID, addr string) {
	d.primaryMux.RLock()
	id, addr = d.primaryDC, d.primaryAddr
	d.primaryMux.RUnlock()
	return
}

func (d *dcInfo) Store(id dcID, addr string) {
	d.primaryMux.Lock()
	d.primaryDC, d.primaryAddr = id, addr
	d.primaryMux.Unlock()
}
