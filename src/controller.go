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
	chmap      map[uint32]chan CReq
	idCounter  uint32
}

var chansize = 1000

/* Ideally this needs to request the devices from another interface */
/** TODO need a connection finding phase **/
func (c *Controller) Initialize(devices int) {
	c.NumDevices = 0
	c.DBuffer = make(chan *Device, 1000)
	c.chmap = make(map[uint32]chan CReq)
	InitializeDeviceManagement()
	// Add to the map and add to the buffer as well...
	for i := 0; i < devices; i++ {
		// Create some dummy device
		// Ideally we want to provision some n devices from switch, etc.
		raddr, err := net.ResolveTCPAddr("tcp", "192.168.7.1:8820")
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
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:7070")
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
		msg, _ := ReadMessage(*conn)
		var msgType uint8 = uint8(msg[4])
		// If OPEN_CONN_1
		if msgType == 38 {
			fmt.Println("OPEN_CONN_1")
			// Instantiate a goroutine
			gid := atomic.AddUint32(&c.idCounter, 1)
			ch := make(chan CReq, chansize)
			c.chmap[gid] = ch
			go c.OperateDeviceOnInstance(gid, msg, ch, *conn)
		} else {
			fmt.Println("Received some cascade instance request")
			msgGid := ReadInt32(msg[5:9])
			fmt.Printf("msgGid: %d\n", msgGid)
			// If we are a OPEN_CONN_2
			c.chmap[msgGid] <- CReq{*conn, msg}
		}
	}
}

/** We instantiate one of these per cascade instance **/
func (c *Controller) OperateDeviceOnInstance(gid uint32, initMsg []byte, ch chan CReq, oc1 net.TCPConn) {
	// Get device
	var err error
	dev := <-c.DBuffer
	conn := dev.GetOC1()
	dev.PID, err = Handshake(oc1, conn, initMsg, gid)
	if err != nil {
		panic(err)
	}
	dev.GID = gid
	go dev.DoForwarding(oc1, conn)

	for {
		newCReq := <-ch
		msgType := uint8(newCReq.Buff[4])
		if msgType == 39 {
			conn = dev.GetOC2()
		} else {
			conn = dev.GetNextConn()
		}
		gid := TranslateGIDPID(&newCReq.Buff, dev.PID)
		_, err = Handshake(newCReq.Conn, conn, newCReq.Buff, gid)
		if err == nil {
			go dev.DoForwarding(newCReq.Conn, conn)
		}
	}
}
