package hasher

import (
	"testing"
	"time"
)

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
	}
}
