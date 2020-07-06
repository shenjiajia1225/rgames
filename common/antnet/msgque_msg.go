package antnet

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

const (
	MsgHeadSize = 8
)

var MaxMsgDataSize uint32 = 1024 * 1024

type MessageHead struct {
	Len uint32
	Id  uint32
}

func (r *MessageHead) Bytes() []byte {
	data := make([]byte, MsgHeadSize)
	phead := (*MessageHead)(unsafe.Pointer(&data[0]))
	phead.Len = r.Len
	phead.Id = r.Id
	return data
}

func (r *MessageHead) FastBytes(data []byte) []byte {
	phead := (*MessageHead)(unsafe.Pointer(&data[0]))
	phead.Len = r.Len
	phead.Id = r.Id
	return data
}

func (r *MessageHead) BytesWithData(wdata []byte) []byte {
	r.Len = uint32(len(wdata))
	data := make([]byte, MsgHeadSize+r.Len)
	binary.BigEndian.PutUint32(data, r.Len)
	binary.BigEndian.PutUint32(data[4:], r.Id)
	if wdata != nil {
		copy(data[MsgHeadSize:], wdata)
	}
	return data
}

func (r *MessageHead) FromBytes(data []byte) error {
	if len(data) < MsgHeadSize {
		return ErrMsgLenTooShort
	}
	r.Len = binary.BigEndian.Uint32(data)
	r.Id = binary.BigEndian.Uint32(data[4:])
	if r.Len > MaxMsgDataSize {
		return ErrMsgLenTooLong
	}
	return nil
}

func (r *MessageHead) String() string {
	return fmt.Sprintf("Len:%v Id:%v", r.Len, r.Id)
}

func NewMessageHead(data []byte) *MessageHead {
	head := &MessageHead{}
	if err := head.FromBytes(data); err != nil {
		return nil
	}
	return head
}

type Message struct {
	Head *MessageHead //消息头，可能为nil
	Data []byte       //消息数据
	User interface{}  //用户自定义数据
}

func (r *Message) Len() uint32 {
	if r.Head != nil {
		return r.Head.Len
	}
	return 0
}

func (r *Message) Bytes() []byte {
	if r.Head != nil {
		if r.Data != nil {
			return r.Head.BytesWithData(r.Data)
		}
		return r.Head.Bytes()
	}
	return r.Data
}

func NewDataMsg(data []byte, id int) *Message {
	return &Message{
		Head: &MessageHead{
			Len: uint32(len(data)),
			Id:  uint32(id),
		},
		Data: data,
	}
}
