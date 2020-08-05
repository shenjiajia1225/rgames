package game1

import (
	//"rgames/common/base"
	"sync"
)

func init() {
}

var (
	once     sync.Once
	instance *MoveManager
)

type MoveManager struct {
	impls map[int64]MoveImpl
}

func MoveMgrInstance() *MoveManager {
	once.Do(func() {
		instance = &MoveManager{
			impls: make(map[int64]MoveImpl),
		}
	})
	return instance
}

func (mm *MoveManager) Add(id int64, mimpl MoveImpl) {
	if _, ok := mm.impls[id]; !ok {
		mm.impls[id] = mimpl
	}
}

func (mm *MoveManager) Remove(id int64) {
	if _, ok := mm.impls[id]; ok {
		delete(mm.impls, id)
	}
}

func (mm *MoveManager) Update(now int64) {
	for _, impl := range mm.impls {
		if impl != nil {
			impl.Update(now)
		}
	}
}
