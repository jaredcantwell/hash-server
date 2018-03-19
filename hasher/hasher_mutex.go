package hasher

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncHasherMutex is an implementation of the AsyncHasher interface
// that uses mutexes as the primary means of synchronization.  For such
// a simple application using channels for synchronization seems like
// overkill.  See AsyncHasherChannel for an implementation using channels.
type AsyncHasherMutex struct {
	hashMutex sync.Mutex
	asyncId   int64 // counter of ids to return to ensure uniqueness
	hashes    map[int64]string

	statsMutex sync.Mutex
	stats      Stats

	wg sync.WaitGroup // Used to wait for all long-running operations to complete on shutdown
}

// NewHasherMutex creates and initializes a new AsyncHasher.
func NewHasherMutex() AsyncHasher {
	var hasher AsyncHasherMutex
	hasher.hashes = make(map[int64]string)
	return &hasher
}

// Compute schedules the supplied password to be hashed asynchronously and
// returns an id that can be supplied to GetAndRemoveHash at a later time to
// retrieve the hash.  For details on the hash, see hasher.Compute.
func (h *AsyncHasherMutex) Compute(password string) int64 {
	// Atomically incrementing is the easiest way to have non-conflicting ids.
	// If security was a concern, we'd want to consider returning a random integer,
	// or even better a long alphanumeric key.
	id := atomic.AddInt64(&h.asyncId, 1)

	h.wg.Add(1)
	go func() {
		// The purpose of this sleep is to simulate a longer running
		// task, so we just sleep.  I considered using time.After along
		// with a channel to cancel the task mid-operation, but instead
		// opted to assume this was a "long" running task that is NOT
		// cancelable.  This means we just have to wait for it to complete
		// when shutting down.
		time.Sleep(5 * time.Second)

		// For stats, we're only interested in the real work, which is the hash.
		// Maybe if the sleep were real work, we would include that too.
		start := time.Now()
		hash := Compute(password)

		h.statsMutex.Lock()
		h.stats.update(time.Since(start))
		h.statsMutex.Unlock()

		h.hashMutex.Lock()
		h.hashes[id] = hash
		h.hashMutex.Unlock()

		h.wg.Done()
	}()

	return id
}

// GetAndRemoveHash returns the hash that was computed in the background for
// the supplied id, and also removes it from our cache.  Therefore,
// this function will only return a hash one time for a given id.
// This id must have been returned from a previous Compute call.
// If the hash is not completed yet, an error will be returned.
func (h *AsyncHasherMutex) GetAndRemoveHash(id int64) (string, error) {
	h.hashMutex.Lock()
	defer h.hashMutex.Unlock()

	// get entry in the map and put it back on the channel
	val, exists := h.hashes[id]
	if !exists {
		return "", errors.New("id not found")
	}

	// After the value is retrieved, remove it from the map.  This is typical
	// behavior for asynchronous operations in order to avoid our map growing
	// boundlessly
	delete(h.hashes, id)
	return val, nil
}

// Stats returns the current statistics about performance of the hash
// computations being performed, including the total number of Compute
// requests and the average time (in milliseconds) to perform the hash
// computation.
func (h *AsyncHasherMutex) Stats() Stats {
	h.statsMutex.Lock()
	defer h.statsMutex.Unlock()

	return h.stats
}

// Drain cleans up the AsyncHasher and waits for all outstanding asynchronous
// hashes to complete in the background (which could take several seconds
// because we're simulating these being an expensive operation).  When Drain
// returns, all resources for the AsyncHasher are in a clean shutdown state.
func (h *AsyncHasherMutex) Drain() {
	h.wg.Wait()
}
