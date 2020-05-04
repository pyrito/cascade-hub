package main

import (
	"fmt"
	"net"
	"sync/atomic"
)

/** Stuff required for each Cascade instance **/
type CReq struct {
	Conn net.TCPConn
	Buff []byte
}

/** One type of controller scheduling property **/
/** TODO: create some kind of interface that accepts different schedulers **/
/** TODO: Put a lock around this stuff **/
type Controller struct {
	NumDevices int
	DBuffer    chan *Device
	CReqChan   chan CReq
	chmap      map[uint32]CReq
	idCounter  uint32
}

var chansize = 1000

/* Ideally this needs to request the devices from another interface */
/** TODO need a connection finding phase **/
func (c *Controller) Initialize(devices int) {
	c.NumDevices = 0
	c.DBuffer = make(chan *Device, 1000)
	c.CReqChan = make(chan CReq, 1000)
	InitializeDeviceManagement()
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
	fmt.Println("Done initializing...")
}

/* Add a device to the controller */
func (c *Controller) AddDevice(raddr *net.TCPAddr) {
	c.NumDevices++
	d := NewDevice(*raddr)
	c.DBuffer <- d
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
		conn, err := l.AcceptTCP()
		if err != nil {
			panic(err)
		}
		msg := ReadMessage(conn)
		var msgType uint8 = uint8(msg[4])
		// If OPEN_CONN_1
		if msgType == 37 {
			// Instantiate a goroutine
			gid := atomic.AddU32(&c.idCounter, 1)
			ch := make(chan CReq, chansize)
			c.chmap[gid] = ch
			go c.OperateDeviceOnInstance(gid, msg, ch, conn)
		} else {
			msgGid := ReadInt32(msg[5:9])
			// If we are a OPEN_CONN_2
			c.chmap[msgGid] <- CReq{conn, msg}
		}
	}
}

/** We instantiate one of these per cascade instance **/
func (c *Controller) OperateDeviceOnInstance(gid uint64, initMsg []byte, ch chan CReq, oc1 net.TCPConn) {
	// Get device
	dev := <-c.DBuffer
	conn := dev.GetOC1()
	dev.PID := Handshake(oc1, conn)
	dev.GID = gid
	go dev.DoForwarding(oc1, conn)

	for {
		newCReq := <-ch
		msgType := uint8(newCreq.Buff[4])
		if msgType == 38 {
			conn = dev.GetOC2()
		} else {
			conn = dev.GetNextConn()
		}
		Handshake(newCReq.Conn, conn)
		go dev.DoForwarding(newCreq.Conn, conn)
	}
}
