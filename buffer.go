package main
import (
	"sync"
)

/******************************************/
/******************************************/
/********       BUFFER LOGIC      *********/
/******************************************/
/******************************************/


type DeviceBuffer struct {
	size int
	numElem int
	mut sync.Mutex
	fullCond *sync.Cond
	emptyCond *sync.Cond
	done bool
	buffer []*Device
}

func (b *DeviceBuffer) Initialize(size int) {
	b.fullCond = sync.NewCond(&b.mut)
	b.emptyCond = sync.NewCond(&b.mut)
	b.size = size
	b.numElem = 0
	b.buffer = make([]*Device, size, size)
	b.done = false
}

func (b *DeviceBuffer) Enqueue(cp *Device) {
	b.mut.Lock()
	defer b.mut.Unlock()
	// Handle case if the queue is already full
	for b.numElem == b.size {
		b.emptyCond.Wait()
	}

	b.buffer[b.numElem] = cp
	b.numElem++
	b.fullCond.Signal()
}

func (b *DeviceBuffer) Dequeue() *Device{
	b.mut.Lock()
	defer b.mut.Unlock()
	// Handle case if there is nothing to remove from the queue
	for b.numElem == 0 {
		b.fullCond.Wait()
	}

	ret := b.buffer[0]
	// Need to check what to do if the device is considered dead
	b.buffer = append(b.buffer[1:], EmptyDevice())
	b.numElem--
	b.emptyCond.Signal()
	return ret
}

func (b *DeviceBuffer) Peek() *Device {
	b.mut.Lock()
	defer b.mut.Unlock()
	ret := b.buffer[0]
	return ret
}