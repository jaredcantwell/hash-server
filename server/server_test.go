package server

import (
	"bytes"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

// Missing Tests
// -------------
// These tests should be added in order to have complete test coverage:
// - Verify that bad port input correctly errors out gracefully
// - Verify that many simulataneous shutdown requests are properly handled and the
//   server still shuts down correctly.
// - Verify correct behavior if there are incoming requests while trying to shutdown.
//   These new requests should be rejected.
// - Verify that the following request error cases are properly detected:
//   - Many combinations of invalid request paths
//   - Invalid methods (GET/POST/DELETE) on valid paths
//   - Requests for results before the background hashing is complete
//   - Requests for results of ids that never existed
//   - Requests for results that have already been retrieved
//   - Verify different password param permutations

func TestStress(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		New(8080).Run()
		wg.Done()
	}()

	// Sleep 2 seconds to let the server startup.. this obviously needs fixed/improved
	time.Sleep(2 * time.Second)

	responses := make(chan string, 100)

	// note: this can scale much higher, but requires changing ulimit and
	// I didn't want to require that to run the tests
	for i := 0; i < 100; i++ {
		go func() {
			resp, err := http.PostForm("http://localhost:8080/hash",
				url.Values{"password": {"angryMonkey"}})

			if err != nil {
				t.Fail()
				return
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			responses <- buf.String()
		}()
	}

	// Sleep enough so that some hashes complete and others are still queued.
	// NOTE: sleep in unittests is generally bad.  I would much prefer that we
	// could test with a more controlled hash function with hooks into the UT,
	// but I didn't have time to implement that.
	time.Sleep(10 * time.Second)

	_, err := http.Post("http://localhost:8080/shutdown", "", nil)
	if err != nil {
		t.Fail()
	}

	wg.Wait()
}
