package hzmj

import (
	"fmt"
	"rgames/common/base"
	"rgames/common/utils"
)

func init() {
}

type Hall struct {
	base.HallBase
	tell int32
}

func Create(gameid int32) base.HallImpl {
	h := &Hall{}
	h.Init(gameid, h, CreateUser, CreateRoom)
	return h
}

func (h *Hall) HandleOtherMessage(um *base.UserMessage) {
	utils.TLog.Warn("HZMJ HallBase::HandleOtherMessage")
	fmt.Printf("tell=%v\n", h.tell)
}
