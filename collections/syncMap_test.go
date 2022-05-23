package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSyncMapLockMistakes(t *testing.T) {
	mp := NewSyncMap[int, int]()
	mp.Put(1, 2)
	val, _ := mp.Get(1)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, mp.Size())
	assert.Equal(t, 1, len(mp.GetSlice()))
	mp.Delete(1)
	assert.Equal(t, 0, mp.Size())
	//Add exec all operations second time - if any lock misspelled it may lock
	mp.Put(1, 3)
	val, _ = mp.Get(1)
	assert.Equal(t, 3, val)
	assert.Equal(t, 1, mp.Size())
	mp.Delete(1)
	assert.Equal(t, 0, mp.Size())
}
