package ddz

import (
	"rgames/common/base"
)

func init() {

}

type Room struct {
	base.RoomBase
}

func CreateRoom(count int32, passwd string) base.RoomImpl {
	r := &Room{}
	r.Init(r, count, passwd)
	return r
}

func (r *Room) OnTimeOut(now int64) {

}

func (r *Room) OnGameStart() {
}

func (r *Room) OnGameEnd() {
}

func (r *Room) OnRoundEnd() {
}

func (r *Room) HandleOtherMessage(um *base.UserMessage) bool {
	return false
}

func (r *Room) CanEnter(user base.UserImpl, passwd string) int32 {
	ret := r.RoomBase.CanEnter(user, passwd)
	return ret
}

func (r *Room) CanLeave(user base.UserImpl, reason int32) int32 {
	ret := r.RoomBase.CanLeave(user, reason)
	return ret
}
