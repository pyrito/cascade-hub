package main

import (
	// "time"
	// "math/rand"
	"net"
	//"io/ioutil"
	// "io"
	// "bufio"
	// "os"
)

var cascade_host string

type CReq struct {
	Conn *net.TCPConn
	Buff []byte
}

/** One type of controller scheduling property **/
/** TODO: create some kind of interface that accepts different schedulers **/
/** TODO: Put a lock around this stuff **/
type Controller struct {
	NumDevices int
	DBuffer    chan *Device
	Dmap       map[net.TCPConn]struct{}
	Cmap       map[net.TCPConn]chan []byte
	CReqChan   chan CReq
}

var cid uint64 = 0

/* Ideally this needs to request the devices from another interface */
/** TODO need a connection finding phase **/
func (c *Controller) Initialize(devices int) {
	c.NumDevices = 0
	c.Dmap = make(map[net.TCPConn]struct{})
	c.Cmap = make(map[net.TCPConn]chan []byte)
	c.CReqChan = make(chan CReq)
	// Add to the map and add to the buffer as well...
	for i := 0; i < devices; i++ {
		// Create some dummy device
		// Ideally we want to provision some n devices from switch, etc.
		raddr, err := net.ResolveTCPAddr("tcp", "192.168.7.1:8800")
		if err != nil {
			panic(err)
		}
		c.AddDevice(raddr)
	}
}

/* Add a device to the controller */
func (c *Controller) AddDevice(raddr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		panic(err)
	}
	devices := c.Dmap
	if _, ok := devices[*conn]; ok {
	} else {
		c.NumDevices += 1
		devices[*conn] = struct{}{}
		d := NewDevice(conn, false, -1)
		c.DBuffer <- d
	}
}

/* Remove a device from the controller */
/** INVARIANT Connection being checked was taken off of enqueue **/
func (c *Controller) RemoveDevice(conn net.TCPConn) {
	available := c.Dmap
	if _, ok := available[conn]; ok {
		delete(available, conn)
	}
}

func (c *Controller) AddCascade(conn net.TCPConn) chan []byte {
	cascades := c.Cmap
	if _, ok := cascades[conn]; ok {
	} else {
		cascades[conn] = make(chan []byte)
		go c.OperateDeviceOnInstance(cascades[conn], conn)
	}
	return cascades[conn]
}

/** INVARIANT Cascade being removed should have been already closed **/
func (c *Controller) RemoveCascade(conn net.TCPConn) {
	available := c.Cmap

	if ch, ok := available[conn]; ok {
		close(ch)
		delete(available, conn)
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
	go c.SendToOwn()

	for {
		// Wait for a connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			panic(err)
		}
		msg := ReadMessage(conn)
		c.CReqChan <- CReq{conn, msg}
	}
}

/** Send things to their own queue. **/
/** TODO Deal with Cascade deaths **/
func (c *Controller) SendToOwn() {
	for {
		creq := <-c.CReqChan
		ch := c.AddCascade(*creq.Conn)
		ch <- creq.Buff
	}
}

/** We instantiate one of these per cascade instance **/
func (c *Controller) OperateDeviceOnInstance(ch chan []byte, conn net.TCPConn) {
	dev := <-c.DBuffer
	for {
		msg := <-ch
		_, err := dev.Conn.Write(msg)
		if err != nil {
			panic(err)
		}
		res := ReadMessage(dev.Conn)
		_, err = conn.Write(res)
		if err != nil {
			panic(err)
		}
	}
}
