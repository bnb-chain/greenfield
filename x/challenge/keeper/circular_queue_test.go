package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCircularQueue(t *testing.T) {
	cq := &circularQueue{
		size:  5,
		items: make([]uint64, 5),
		front: -1,
	}

	// enqueue some items
	cq.enqueue(1)
	cq.enqueue(2)
	cq.enqueue(3)
	require.Equal(t, []uint64{1, 2, 3}, cq.retrieveAll())

	// enqueue more to override older ones
	cq.enqueue(4)
	cq.enqueue(5)
	cq.enqueue(6)
	require.Equal(t, []uint64{2, 3, 4, 5, 6}, cq.retrieveAll())

	// shrink the queue, which will remove the older ones
	cq.resize(3)
	require.Equal(t, []uint64{4, 5, 6}, cq.retrieveAll())

	// enlarge the queue
	fmt.Println(cq)
	cq.resize(5)
	fmt.Println(cq)
	cq.enqueue(7)
	cq.enqueue(8)
	require.Equal(t, []uint64{4, 5, 6, 7, 8}, cq.retrieveAll())
	cq.enqueue(9)
	cq.enqueue(10)
	require.Equal(t, []uint64{6, 7, 8, 9, 10}, cq.retrieveAll())
}
