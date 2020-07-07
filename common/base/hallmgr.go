package base

import (
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	"rgames/common/utils"
	"sync"
)

func init() {
}

var (
	once     sync.Once
	instance *HallMgr
)

func GetHallMgr() *HallMgr {
	once.Do(func() {
		instance = &HallMgr{
			createCallbacks: make(map[int32]func(int32) HallImpl),
			Halls:           cmap.New(),
			Conn2GameMap:    cmap.New(),
		}
	})
	return instance
}

type HallMgr struct {
	createCallbacks map[int32]func(int32) HallImpl
	Halls           cmap.ConcurrentMap
	Conn2GameMap    cmap.ConcurrentMap
}

func (hm *HallMgr) Register(gameid int32, fc func(int32) HallImpl) {
	hm.createCallbacks[gameid] = fc
	hm.CreateHall(gameid)
	utils.TLog.Info(fmt.Sprintf("Register game[%v]", gameid))
}

func (hm *HallMgr) UnRegisterAll() {
	hm.Halls.IterCb(func(key string, v interface{}) {
		h := v.(HallImpl)
		if h != nil {
			utils.TLog.Info(fmt.Sprintf("UnRegister game[%v]", h.GameID()))
			h.Fini()
		}
	})
}

func (hm *HallMgr) CreateHall(gameid int32) {
	cb, ok := hm.createCallbacks[gameid]
	if ok {
		impl := cb(gameid)
		hm.Add(gameid, impl)
	} else {
		utils.TLog.Error(fmt.Sprintf("CreateHall game[%v]", gameid))
	}
}

func (hm *HallMgr) Add(gameid int32, impl HallImpl) {
	_, ok := hm.Halls.Get(string(gameid))
	if !ok {
		hm.Halls.Set(string(gameid), impl)
	}
	utils.TLog.Info(fmt.Sprintf("Add hall[%v] ok[%v]", gameid, ok))
}

func (hm *HallMgr) Get(gameid int32) HallImpl {
	utils.TLog.Debug(fmt.Sprintf("Get hall[%v]", gameid))
	v, ok := hm.Halls.Get(string(gameid))
	if ok {
		return v.(HallImpl)
	} else {
		return nil
	}
}

func (hm *HallMgr) AddConn2Game(connid int64, gameid int32) {
	hm.Conn2GameMap.Set(string(connid), gameid)
}

func (hm *HallMgr) DelConn2Game(connid int64) {
	hm.Conn2GameMap.Remove(string(connid))
}

func (hm *HallMgr) GetConn2Game(connid int64) int32 {
	gameid, ok := hm.Conn2GameMap.Get(string(connid))
	if ok {
		return gameid.(int32)
	} else {
		return 0
	}
}
