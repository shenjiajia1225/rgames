package ddz

import (
	"fmt"
	"rgames/common/base"
	"rgames/common/utils"
)

func init() {
}

type Hall struct {
	base.HallBase
	num int32
}

func Create(gameid int32) base.HallImpl {
	h := &Hall{}
	h.Init(gameid, h, CreateUser, CreateRoom)
	return h
}

func (h *Hall) HandleOtherMessage(um *base.UserMessage) {
	utils.TLog.Warn("DDZ HallBase::HandleOtherMessage")
	fmt.Printf("num=%v\n", h.num)
}
