package utils

import (
	"math/rand"
	"sync"
	"time"
)

var (
	once     sync.Once
	instance *RoomIdMgr
)

type RoomIdMgr struct {
	idList    []int
	syncMutex sync.RWMutex
	start     int
	end       int
	hasInit   bool
}

func GetRoomIdMgr() *RoomIdMgr {
	once.Do(func() {
		instance = &RoomIdMgr{}
	})

	return instance
}

func InitRoomIdMgr(start, end int) {
	mgr := GetRoomIdMgr()
	mgr.Reset(start, end)
}

func (m *RoomIdMgr) Generate() int {
	m.syncMutex.Lock()
	defer m.syncMutex.Unlock()

	if len(m.idList) <= 0 {
		return 0
	}
	id := m.idList[0]
	m.idList = m.idList[1:]

	return id
}

func (m *RoomIdMgr) Release(roomid int) {
	m.syncMutex.Lock()
	defer m.syncMutex.Unlock()

	m.idList = append(m.idList, roomid)
}

func (m *RoomIdMgr) Reset(start, end int) {
	m.hasInit = true
	m.start = start
	m.end = end
	m.idList = m.randList(start, end)
}

func (m *RoomIdMgr) randList(min, max int) []int {
	if max < min {
		min, max = max, min
	}
	length := max - min + 1
	rand.Seed(time.Now().UnixNano())
	list := rand.Perm(length)
	for index, _ := range list {
		list[index] += min
	}
	return list
}
