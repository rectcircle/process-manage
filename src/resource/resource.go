package resource

import (
	"container/list"
)

type RIDType string

type RCB struct {
	RID                   RIDType
	MaxResCnt, FreeResCnt int
	WaitingList           *list.List //等待此资源的PCB
}

func NewRCB(RID RIDType, initResCnt int) *RCB {
	rcb := new(RCB)
	rcb.RID = RID
	rcb.MaxResCnt = initResCnt
	rcb.FreeResCnt = initResCnt
	rcb.WaitingList = list.New()
	return rcb
}
