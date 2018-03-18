package hasher

import (
	"testing"
	"time"
)

// Missing Tests
// -------------
// These tests should be added in order to have complete test coverage:
// - Many outstanding Compute calls at once
// - GetAndRemoveHash should return an error if hash isn't computed yet
// - GetAndRemoveHash should return an error if an invalid hash is provided
// - GetAndRemoveHash should return an error if a hash was already retrieved
// - Verify that Stats properly updates totals and averages
// - Verify many concurrent calls to Stats
// - Stress test many calls to Stats at once (or in quick succession)

func TestHasher(t *testing.T) {
	h := New()

	id := h.Compute("angryMonkey")
	_, err := h.GetAndRemoveHash(id)
	if err == nil {
		t.Fail()
	}

	for {
		time.Sleep(time.Second)
		hash, err := h.GetAndRemoveHash(id)
		if err != nil {
			continue
		}

		if hash != Compute("angryMonkey") {
			t.Fail()
		}

		break
	}

	h.Drain()
}
