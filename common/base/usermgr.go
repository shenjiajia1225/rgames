package base

import (
	"fmt"
	"rgames/common/utils"
)

func init() {

}

type UserMgr struct {
	Users   map[int64]UserImpl
	UserIdx map[int64]int64
}

func CreateUserMgr() *UserMgr {
	um := &UserMgr{
		Users:   make(map[int64]UserImpl),
		UserIdx: make(map[int64]int64),
	}
	return um
}

func (um *UserMgr) Find(connid int64) UserImpl {
	uid, ok := um.UserIdx[connid]
	if !ok {
		return nil
	}

	impl, ok := um.Users[uid]
	if !ok {
		delete(um.UserIdx, connid)
		return nil
	}
	return impl
}

func (um *UserMgr) FindByUserid(userid int64) UserImpl {
	impl, ok := um.Users[userid]
	if !ok {
		return nil
	}
	return impl
}

func (um *UserMgr) Add(impl UserImpl) {
	connid := impl.ConnId()
	userid := impl.UserId()
	_, ok1 := um.UserIdx[connid]
	_, ok2 := um.Users[userid]
	if !ok1 && !ok2 {
		um.UserIdx[connid] = userid
		um.Users[userid] = impl
		utils.TLog.Info(fmt.Sprintf("Add user[%v][%v]", connid, userid))
	} else {
		utils.TLog.Warn(fmt.Sprintf("Add user[%v][%v] exist ok[%v][%v]", connid, userid, ok1, ok2))
		um.UserIdx[connid] = userid
		um.Users[userid] = impl
	}
}

func (um *UserMgr) AddIdx(connid, userid int64) {
	_, ok := um.UserIdx[connid]
	um.UserIdx[connid] = userid
	utils.TLog.Debug(fmt.Sprintf("AddIdx user[%v][%v] exist ok[%v]", connid, userid, ok))
}

func (um *UserMgr) Del(userid int64) {
	u, ok := um.Users[userid]
	if !ok {
		utils.TLog.Error(fmt.Sprintf("Del user[%v]", userid))
		return
	}

	delete(um.Users, userid)
	connId := u.ConnId()
	_, ok = um.UserIdx[connId]
	if !ok {
		utils.TLog.Error(fmt.Sprintf("Del2 user[%v] connid[%v]", userid, connId))
		return
	}
	delete(um.UserIdx, connId)
	utils.TLog.Info(fmt.Sprintf("Del user[%v] connid[%v]", userid, connId))
}

func (um *UserMgr) DelIdx(connid int64) {
	_, ok := um.UserIdx[connid]
	delete(um.UserIdx, connid)
	utils.TLog.Debug(fmt.Sprintf("DelIdx connid[%v] ok[%v]", connid, ok))
}
