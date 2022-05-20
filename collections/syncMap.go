package collections

import "sync"

//SyncMap classic synchronized map through RWMutex implementation
type SyncMap[T1 comparable, T2 any] struct {
	mx sync.RWMutex
	m  map[T1]T2
}

type MapEntry[T1 comparable, T2 any] struct {
	Key   T1
	Value T2
}

//Put or replace key in the map
func (sm *SyncMap[T1, T2]) Put(key T1, value T2) {
	sm.mx.Lock()
	defer sm.mx.Unlock()
	sm.m[key] = value
}

//Delete key from map with write locking
func (sm *SyncMap[T1, T2]) Delete(key T1) {
	sm.mx.Lock()
	defer sm.mx.Unlock()
	delete(sm.m, key)
}

//Get value from map. Returns bool to check is value exists
func (sm *SyncMap[T1, T2]) Get(key T1) (T2, bool) {
	sm.mx.RLock()
	defer sm.mx.RUnlock()
	val, ok := sm.m[key]
	return val, ok
}

//Size returns len from internal map result
func (sm *SyncMap[T1, T2]) Size() int {
	sm.mx.RLock()
	defer sm.mx.RUnlock()
	return len(sm.m)
}

//GetSlice returns key-value entry slice of map
func (sm *SyncMap[T1, T2]) GetSlice() []*MapEntry[T1, T2] {
	sm.mx.RLock()
	defer sm.mx.RUnlock()
	sl := make([]*MapEntry[T1, T2], 0, len(sm.m))
	for key, val := range sm.m {
		sl = append(sl, &MapEntry[T1, T2]{key, val})
	}
	return sl
}

//Clear clears all entries by recreating internal map
func (sm *SyncMap[T1, T2]) Clear() {
	sm.mx.Lock()
	defer sm.mx.Unlock()
	sm.m = make(map[T1]T2, 0)
}

func NewSyncMap[T1 comparable, T2 any]() SyncMap[T1, T2] {
	return SyncMap[T1, T2]{
		m: make(map[T1]T2, 0),
	}
}
