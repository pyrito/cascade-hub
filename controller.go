package main
import (
	"fmt"
)

/** One type of controller scheduling property **/
/** TODO: create some kind of interface that accepts different schedulers **/
type Controller struct {
	ID int
	NumDevices int
	DevicesAvailable map[int]*Device
	Buffer DeviceBuffer
}

/* Ideally this needs to request the devices from another interface */
func (c *Controller) Initialize(devices int) {
	// TODO: give proper ID
	c.ID = 0
	c.NumDevices = devices
	c.DevicesAvailable = make(map[int]*Device)
	c.Buffer.Initialize(devices)
	// Add to the map and add to the buffer as well...
	for i := 0; i < devices; i++ {
		// Create some dummy device
		// Ideally we want to provision some n devices from switch, etc.
		d := NewDevice("192.168.1.1", i, false, false, -1)
		c.Add(i, d)
		c.Buffer.Enqueue(d)
	}
}

/* Add a device to the controller */
func (c *Controller) Add(id int, d *Device) {
	devices := c.DevicesAvailable
	if _, ok := devices[id]; ok {
		fmt.Println("Rewriting existing value")
	} else {
		devices[id] = d
	}
}

/* Remove a device from the controller */
func (c *Controller) Remove(id int) {
	available := c.DevicesAvailable
	if device, ok := available[id]; ok {
		device.Removed = true
		delete(available, id)
	}
}

/* Send the work to some provisioned FPGA */
func (c *Controller) Execute(filePath string) int {
	var d *Device = c.Buffer.Dequeue()
	// Add the SSH logic in here accordingly
	// Think about security or something probably
	fmt.Printf("d.id: %d\n", d.ID)
	return 0
}