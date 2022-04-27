package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPopulating(t *testing.T) {
	lst := NewTList[int](time.Second)
	nw := time.Now()
	lst.Append(1, nw)
	assert.NotNil(t, lst.First(), "First element must not be nil")
	assert.NotNil(t, lst.Last(), "Last element must not be nil")
	lst.Append(2, nw)
	lst.Append(3, nw)
	assert.Equal(t, uint(3), lst.GetSize(), "List must have actual size")

	upd := nw.Add(time.Minute)
	lst.Append(4, upd)
	assert.Equal(t, uint(1), lst.GetSize(), "After removing elements list must have actual size")
	assert.NotNil(t, lst.First(), "First element must not be nil")
	assert.NotNil(t, lst.Last(), "Last element must not be nil")
	assert.Nil(t, lst.First().Next(), "Next element must be nil when list has only one element")
	assert.Equal(t, lst.First(), lst.Last(), "When one element in list - first and last must be the same")
	lst.Append(5, upd)
	lst.Append(6, upd)
	assert.Equal(t, uint(3), lst.GetSize(), "List must have actual size after repopulating")
}
