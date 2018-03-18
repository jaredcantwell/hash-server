package server

import "testing"

var failInputs = []struct {
	path   string
	prefix string
}{
	{"/some/path", "/some"},
	{"/some/path", "/some/"},
	{"/some/path/", "/"},
	{"/some/path/", "/some/path/"},
}

// TestParse verifies basic failure functionality for the helper method
// that parses the id path parameter out of the URL.
func TestParseFailures(t *testing.T) {
	for _, input := range failInputs {
		_, err := parsePathParamInt(input.path, input.prefix)
		if err == nil {
			t.Fail()
		}
	}
}

// TestParseSuccess verifies that a path parameter is correctly parsed
// out of a path.
func TestParseSuccess(t *testing.T) {
	id, err := parsePathParamInt("/path/345", "/path/")
	if err != nil || id != 345 {
		t.Fail()
	}
}
