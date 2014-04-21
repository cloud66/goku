package main

import (
	"testing"
)

func TestNotify(t *testing.T) {
	h := HttpNotifier{Endpoint: "http://httpbin.org/post"}

	p := Process{
		Name:      "TestStart",
		Directory: "/bin",
		Command:   "sleep",
		Args:      []string{"1"},
	}

	n := fromProcess(&p)

	_, err := h.notify(n)
	if err != nil {
		t.Error(err)
	}
}
