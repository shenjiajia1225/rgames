package base

import (
	"fmt"
	"rgames/common/utils"
)

func init() {

}

type RoomMgr struct {
	Rooms map[int64]RoomImpl
}

func CreateRoomMgr() *RoomMgr {
	rm := &RoomMgr{
		Rooms: make(map[int64]RoomImpl),
	}
	return rm
}

func (rm *RoomMgr) Find(roomid int64) RoomImpl {
	r, ok := rm.Rooms[roomid]
	if !ok {
		return nil
	}
	return r
}

func (rm *RoomMgr) Add(impl RoomImpl) {
	roomid := impl.RoomId()
	_, ok := rm.Rooms[roomid]
	if !ok {
		rm.Rooms[roomid] = impl
		utils.TLog.Info(fmt.Sprintf("Add room[%v]", roomid))
	} else {
		utils.TLog.Warn(fmt.Sprintf("Add room[%v] exist", roomid))
	}
}

func (rm *RoomMgr) Del(roomid int64) {
	_, ok := rm.Rooms[roomid]
	if !ok {
		utils.TLog.Error(fmt.Sprintf("Del room[%v]", roomid))
		return
	}

	delete(rm.Rooms, roomid)
	utils.TLog.Info(fmt.Sprintf("Del room[%v]", roomid))
}
