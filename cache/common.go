package cache

import (
	"time"
)

func gc(c Cache, stop chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		// delete expired item on every tick
		case <-ticker.C:
			c.cleanExpired()

		// stop ticker when exist
		case <-stop:
			return

		}

	}

}

func unixNow() int64 {
	return time.Now().UTC().Unix()
}
