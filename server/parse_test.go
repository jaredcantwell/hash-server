package server

import "testing"

// TODO: refactor this more
func TestParse(t *testing.T) {
	_, err := parsePathParamInt("/some/path/", "/some")
	if err == nil {
		t.Fail()
	}

	_, err = parsePathParamInt("/some/path/", "/some/")
	if err == nil {
		t.Fail()
	}

	_, err = parsePathParamInt("/some/path/", "/")
	if err == nil {
		t.Fail()
	}

	_, err = parsePathParamInt("/some/path/", "/some/path/")
	if err == nil {
		t.Fail()
	}

	id, err := parsePathParamInt("/path/345", "/path/")
	if err != nil {
		t.Fail()
	}
	if id != 345 {
		t.Fail()
	}
}
