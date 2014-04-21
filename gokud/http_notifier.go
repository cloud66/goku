package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"code.google.com/p/go-uuid/uuid"
	"github.com/golang/glog"
)

var (
	USER_AGENT string = "goku/" + VERSION + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
)

type HttpNotifier struct {
	Endpoint string

	client *http.Client
}

func (n *HttpNotifier) notify(process *Notification) (string, error) {
	httpClient := n.client
	if httpClient == nil {
		n.client = http.DefaultClient
	}

	var rbody io.Reader

	j, err := json.Marshal(process)
	if err != nil {
		glog.Error(err)
	}
	rbody = bytes.NewReader(j)

	req, err := http.NewRequest("POST", n.Endpoint, rbody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Request-Id", uuid.New())
	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	res, err := n.client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
