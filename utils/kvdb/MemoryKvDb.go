package kvdb

import (
	"errors"
	"sync"
	"time"
)

type MemoryKvDb struct {
	keys       map[string]any
	expireTime map[string]int64
	mu         sync.RWMutex
}

func (m *MemoryKvDb) Get(key string) (val any, err error) {
	m.mu.RLock()
	val, isExist := m.keys[key]
	if !isExist {
		m.mu.RUnlock()
		return nil, errors.New("invalid key")
	}
	expireTime, isExist := m.expireTime[key]
	if !isExist {
		m.mu.RUnlock()
		return val, nil
	}
	if time.Now().After(time.Unix(expireTime, 0)) {
		m.mu.RUnlock()
		return nil, errors.New("invalid key")
	}
	m.mu.RUnlock()
	return val, nil
}

func (m *MemoryKvDb) GetInt(key string) (i int32, err error) {
	v, err := m.Get(key)
	if err != nil {
		return 0, err
	}
	if i, ok := v.(int32); ok {
		return i, nil
	}
	return 0, errors.New("invalid type")
}

func (m *MemoryKvDb) GetInt64(key string) (i64 int64, err error) {
	v, err := m.Get(key)
	if err != nil {
		return 0, err
	}
	if i64, ok := v.(int64); ok {
		return i64, nil
	}
	return 0, errors.New("invalid type")
}

func (m *MemoryKvDb) GetString(key string) (str string, err error) {
	v, err := m.Get(key)
	if err != nil {
		return "", err
	}
	if str, ok := v.(string); ok {
		return str, nil
	}
	return "", errors.New("invalid type")
}

func (m *MemoryKvDb) GetBool(key string) (b bool, err error) {
	v, err := m.Get(key)
	if err != nil {
		return false, err
	}
	if b, ok := v.(bool); ok {
		return b, nil
	}
	return false, errors.New("invalid type")
}

func (m *MemoryKvDb) Set(key string, value any) (err error) {
	m.mu.Lock()
	m.keys[key] = value
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) SetInt(key string, value int) (err error) {
	m.mu.Lock()
	m.keys[key] = value
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) SetInt64(key string, value int64) (err error) {
	m.mu.Lock()
	m.keys[key] = value
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) SetString(key string, value string) (err error) {
	m.mu.Lock()
	m.keys[key] = value
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) SetBool(key string, value bool) (err error) {
	m.mu.Lock()
	m.keys[key] = value
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) Delete(key string) (err error) {
	m.mu.Lock()
	_, isExist := m.keys[key]
	if !isExist {
		m.mu.Unlock()
		return errors.New("invalid key")
	}
	delete(m.keys, key)
	_, isExist = m.expireTime[key]
	if isExist {
		delete(m.expireTime, key)
	}
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) Expire(key string, expireTime int64) (err error) {
	return m.ExpireAt(key, time.Now().Unix()+expireTime)
}

func (m *MemoryKvDb) ExpireAt(key string, expireTimestamp int64) (err error) {
	m.mu.Lock()
	_, isExist := m.keys[key]
	if !isExist {
		m.mu.Unlock()
		return errors.New("invalid key")
	}
	m.expireTime[key] = expireTimestamp
	m.mu.Unlock()
	return nil
}

func (m *MemoryKvDb) Persist(key string) (err error) {
	m.mu.Lock()
	_, isExist := m.keys[key]
	if !isExist {
		m.mu.Unlock()
		return errors.New("invalid key")
	}
	_, isExist = m.expireTime[key]
	if isExist {
		delete(m.expireTime, key)
	}
	m.mu.Unlock()
	return nil
}

func CreateMemoryKvDb() MemoryKvDb {
	return MemoryKvDb{expireTime: make(map[string]int64), keys: make(map[string]any)}
}
