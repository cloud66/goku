package main

import (
	"time"

	"github.com/golang/glog"
)

type Notification struct {
	Uid          string
	Name         string
	CallbackId   string
	Tags         []string
	Pid          int
	LastActionAt time.Time
	LastStatus   int
	StatusCode   int
}

var httpNotifier *HttpNotifier

func startNotifier(events <-chan *Process) {
	for {
		select {
		case process := <-events:
			// do we need notification?
			if httpNotifier != nil && process.lastStatus != process.statusCode {
				// convert to notification
				n := fromProcess(process)
				httpNotifier.notify(n)
				glog.V(Detail).Infof("Sending notification for %s (%v) to %s", process.Name, process.Uid, httpNotifier.Endpoint)
			}
		case <-time.After(time.Second):
			// nop
		}
	}
}

func fromProcess(c *Process) *Notification {
	return &Notification{
		Uid:          c.Uid,
		Name:         c.Name,
		CallbackId:   c.CallbackId,
		Tags:         c.Tags,
		Pid:          c.Pid,
		LastActionAt: c.LastActionAt,
		LastStatus:   c.lastStatus,
		StatusCode:   c.statusCode,
	}
}
