package main
import (
	"fmt"
	// "time"
	// "math/rand"
	"net"
	//"io/ioutil"
	// "io"
	// "bufio"
	// "os"
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
		d := NewDevice("192.168.7.1:8800", i, false, false, -1)
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

/* Built off the example Golang code */
func (c *Controller) ListenToCascade() {
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9090")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		msg := ReadMessage(conn)
		c.Execute(msg)
		fmt.Printf("after reading messages\n")
	}
}

/* Send the work to some provisioned FPGA */
func (c *Controller) Execute(msg []byte) int {
	var d *Device = c.Buffer.Dequeue()

	remoteAddr, err := net.ResolveTCPAddr("tcp", d.IPAddress)
    if err != nil {
        panic(err)
	}
	
	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	n, err := conn.Write(msg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("written bytes: %d\n", n)
	// msg1 := ReadMessage(conn)
	// fmt.Printf("msg len: %d\n", len(msg1))


	// rand.Seed(time.Now().UnixNano())
	// num := rand.Intn(15000)
	// time.Sleep(time.Duration(num) * time.Millisecond)

	fmt.Println("Task finished successfully!")
	
	return 0
}

