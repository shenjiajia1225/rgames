package antnet

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/vmihailenco/msgpack"
)

func JsonUnPack(data []byte, msg interface{}) error {
	if data == nil || msg == nil {
		return ErrJsonUnPack
	}

	err := json.Unmarshal(data, msg)
	if err != nil {
		return ErrJsonUnPack
	}
	return nil
}

func JsonPack(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, ErrJsonPack
	}

	data, err := json.Marshal(msg)
	if err != nil {
		LogInfo("")
		return nil, ErrJsonPack
	}

	return data, nil
}

func MsgPackUnPack(data []byte, msg interface{}) error {
	err := msgpack.Unmarshal(data, msg)
	return err
}

func MsgPackPack(msg interface{}) ([]byte, error) {
	data, err := msgpack.Marshal(msg)
	return data, err
}

func PBUnPack(data []byte, msg interface{}) error {
	if data == nil || msg == nil {
		return ErrPBUnPack
	}

	err := proto.Unmarshal(data, msg.(proto.Message))
	if err != nil {
		return ErrPBUnPack
	}
	return nil
}

func PBPack(msg interface{}) ([]byte, error) {
	if msg == nil {
		return nil, ErrPBPack
	}

	data, err := proto.Marshal(msg.(proto.Message))
	if err != nil {
		LogInfo("")
	}

	return data, nil
}
