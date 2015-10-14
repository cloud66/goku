package models

import (
	"time"
)

type CtrlProcess struct {
	Uid					 string
	Pid					 int
	LastActionAt	time.Time
	TimeStamp		 int64
	Status 			 StatusTuple
}
