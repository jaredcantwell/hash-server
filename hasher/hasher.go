// Package hasher implements an asynchronous hash computation.
//
// Hashing can be quite expensive (up to 5 seconds!), and a caller may not want
// to wait that long synchronously.  The AsyncHasher provides a Compute method
// that allows the caller to request that a password be hashed in the background,
// and returns an id that can be used for later retrieval.  This is intended to be
// used as part of a web application that requires asynchronous polling for
// long-running operations.  If this were to be used internally to go code, returning
// a channel instead of an id would be more appropriate to eliminate polling.
package hasher

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncHasher performs expensive hashing operations in the background and
// provides an interface for the user to retrieve computed hashes at a later
// time asynchronously.
type AsyncHasher interface {
	Compute(password string) int64
	GetAndRemoveHash(id int64) (string, error)
	Stats() Stats
	Drain()
}

// AsyncHasherChannel is an implementation of the AsyncHasher interface
// that uses channels as the primary means of synchronization.  No mutexes
// are used in an attempt to "idomatic" Go.  See AsyncHasherMutex for an
// implementation using mutexes.
type AsyncHasherChannel struct {
	asyncId         int64              // atomic counter of ids to return to ensure uniqueness
	hashPutChan     chan hashPair      // Communicate that a new hash should be cached
	hashRequestChan chan hashRequest   // Communicate a request to retrieve a hash
	statUpdateChan  chan time.Duration // Communicate that an op has completed
	statsChan       chan Stats         // Used to request the latest stats
	shutdown        chan interface{}   // Used to signal shutdown to the event loop
	wg              sync.WaitGroup     // Used to wait for all long-running operations to complete on shutdown
}

// NewHasherChannel creates and initializes a new AsyncHasher.
func NewHasherChannel() AsyncHasher {
	var hasher AsyncHasherChannel
	hasher.hashPutChan = make(chan hashPair, 100)
	hasher.hashRequestChan = make(chan hashRequest, 100)
	hasher.statUpdateChan = make(chan time.Duration, 100)
	hasher.statsChan = make(chan Stats)
	hasher.shutdown = make(chan interface{})

	go hasher.eventLoop()

	return &hasher
}

// Compute schedules the supplied password to be hashed asynchronously and
// returns an id that can be supplied to GetAndRemoveHash at a later time to
// retrieve the hash.  For details on the hash, see hasher.Compute.
func (h *AsyncHasherChannel) Compute(password string) int64 {
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
		h.statUpdateChan <- time.Since(start)

		h.hashPutChan <- hashPair{id, hash}
		h.wg.Done()
	}()

	return id
}

// GetAndRemoveHash returns the hash that was computed in the background for
// the supplied id, and also removes it from our cache.  Therefore,
// this function will only return a hash one time for a given id.
// This id must have been returned from a previous Compute call.
// If the hash is not completed yet, an error will be returned.
func (h *AsyncHasherChannel) GetAndRemoveHash(id int64) (string, error) {
	// Now post a request for the hash for the specified id
	respChan := make(chan hashResponse)
	h.hashRequestChan <- hashRequest{id, respChan}

	// Wait for the response to come back on the channel
	resp := <-respChan

	return resp.hash, resp.err
}

// Stats returns the current statistics about performance of the hash
// computations being performed, including the total number of Compute
// requests and the average time (in milliseconds) to perform the hash
// computation.
func (h *AsyncHasherChannel) Stats() Stats {
	return <-h.statsChan
}

// Drain cleans up the AsyncHasher and waits for all outstanding asynchronous
// hashes to complete in the background (which could take several seconds
// because we're simulating these being an expensive operation).  When Drain
// returns, all resources for the AsyncHasher are in a clean shutdown state.
func (h *AsyncHasherChannel) Drain() {
	h.shutdown <- nil
	h.wg.Wait()
}

// Compute performs a sha512 has on the supplied string and returns the
// base64 encoding the resulting hash.  This is a synchronous operation
// and will complete inline.
func Compute(in string) string {
	sha_512 := sha512.New()
	sha_512.Write([]byte(in))
	out := sha_512.Sum(nil)

	sEnc := base64.StdEncoding.EncodeToString(out)

	return sEnc
}

// eventLoop is where all the heavy synchronization happens.  Since multiple
// callers will be attempting to access data from our map of hashes and the
// common stats value, this event loop uses channels to synchronize all access
// such that this is the only thread touching the map of hashes or the central
// stats value (they are local to this function).
func (h *AsyncHasherChannel) eventLoop() {
	h.wg.Add(1)

	var hashes = make(map[int64]string)
	var stats Stats

loop:
	for {
		select {
		// A hash computation has completed and is adding into the map
		case pair := <-h.hashPutChan:
			hashes[pair.id] = pair.hash
			// A user is requesting the hash for an id
		case req := <-h.hashRequestChan:
			// get entry in the map and put it back on the channel
			val, exists := hashes[req.id]
			if !exists {
				req.resp <- hashResponse{"", errors.New("id not found")}
				break
			}

			// After the value is retrieved, remove it from the map.  This is typical
			// behavior for asynchronous operations in order to avoid our map growing
			// boundlessly
			delete(hashes, req.id)
			req.resp <- hashResponse{val, nil}
			// A user is requesting the latest stats
		case h.statsChan <- stats:
			// The hash computation has completed and is reporting how long it took
		case elapsed := <-h.statUpdateChan:
			stats.update(elapsed)
			// Drain has been called and its time to exit this loop
		case <-h.shutdown:
			break loop
		}
	}

	h.wg.Done()
}

// hashPair represents a new entry to be added into the map of hashes
type hashPair struct {
	id   int64
	hash string
}

// hashRequest represents a user request to retrieve a hash for id
type hashRequest struct {
	id   int64             // The id for the hash to be retrieved
	resp chan hashResponse // A channel to send the response back to the caller
}

// hashResponse is sent back from the event loop to the requesting function
type hashResponse struct {
	hash string // If no error, the requested hash
	err  error  // Error indicating hash retreival failed
}
