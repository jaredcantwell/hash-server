package hasher

import "testing"

// TestCompute verifies the most basic hashing building block-- hasher.Compute
func TestCompute(t *testing.T) {
	if Compute("angryMonkey") != "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==" {
		t.Fail()
	}
}
