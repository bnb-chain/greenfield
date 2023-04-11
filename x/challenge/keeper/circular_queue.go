package keeper

// emptyElement to indicate the empty element in circularQueue.
const emptyElement = 0

// circularQueue is a circular queue to store the latest fixed number of attested challenge ids.
type circularQueue struct {
	size  uint64
	items []uint64
	front int64
}

// enqueue will add an item to the queue.
func (c *circularQueue) enqueue(item uint64) {
	c.front = (c.front + 1) % int64(c.size)
	c.items[c.front] = item
}

// retrieveAll will retrieve all elements from the queue.
func (c *circularQueue) retrieveAll() []uint64 {
	if c.front == -1 {
		return []uint64{}
	}

	result := []uint64{}
	current := c.front
	for {
		current = (current + 1) % int64(c.size)
		if c.items[current] != emptyElement {
			result = append(result, c.items[current])
		}

		if current == c.front {
			break
		}
	}
	return result
}

// resize will update the size of the queue.
func (c *circularQueue) resize(newSize uint64) {
	if newSize == c.size {
		return
	}
	newC := circularQueue{
		size:  newSize,
		items: make([]uint64, newSize),
		front: -1,
	}
	all := c.retrieveAll()
	for _, item := range all {
		newC.enqueue(item)
	}
	*c = newC
}
