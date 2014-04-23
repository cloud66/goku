package main

import (
	"testing"
)

func TestLookupUserId(t *testing.T) {
	id, err := lookupUserId("root")
	if err != nil {
		t.Error(err)
	}
	if id != 0 {
		t.Error(id)
	}
}
