package datastore

import (
	"time"

	"github.com/mhsantos/redis-server/internal/protocol"
)

var (
	store map[string]Value = make(map[string]Value)
)

type Value struct {
	value  protocol.DataType
	expire int64
}

func Get(key string) (protocol.DataType, bool) {
	val, ok := store[key]
	if !ok {
		return nil, false
	}
	if val.IsExpired() {
		delete(store, key)
		return nil, false
	}
	return val.value, ok
}

func Set(key string, value protocol.DataType) {
	val := Value{
		value: value,
	}
	store[key] = val
}

func Delete(key string) bool {
	if _, ok := Get(key); ok {
		delete(store, key)
		return true
	}
	return false
}

func SetWithExpire(key string, value protocol.DataType, expire int64) {
	val := Value{
		value:  value,
		expire: expire,
	}
	store[key] = val
}

func (v Value) IsExpireSet() bool {
	return v.expire > 0
}

func (v Value) IsExpired() bool {
	if v.expire == 0 {
		return false
	}
	return time.Now().Unix() > v.expire
}

func GetWithExpire(key string) (protocol.DataType, int64, bool) {
	val, ok := store[key]
	if !ok {
		return nil, 0, false
	}
	if val.IsExpired() {
		delete(store, key)
		return nil, 0, false
	}
	return val.value, val.expire, true
}
