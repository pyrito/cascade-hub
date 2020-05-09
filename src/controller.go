package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// CReq is used to house a connection and first message
type CReq struct {
	Conn net.TCPConn
	Buff []byte
}

// TReq is used to pass along timings
type TReq struct {
	Conn net.TCPConn
	Time time.Time
}

var timeConn chan TReq

//TimeMesg is a channel used to time message pass over
var TimeMesg chan TReq

// Controller dictates scheduling property
//TODO: create some kind of interface that accepts different schedulers
//TODO: Put a lock around this stuff **/
type Controller struct {
	NumDevices int
	DBuffer    chan *Device
	chmap      map[uint32]chan CReq
	idCounter  uint32
}

var chansize = 1000

// Initialize needs to request the devices from another interface
// TODO need a connection finding phase
func (c *Controller) Initialize(devices int) {
	c.NumDevices = 0
	c.DBuffer = make(chan *Device, 1000)
	c.chmap = make(map[uint32]chan CReq)
	InitializeDeviceManagement()

	//Timing
	timeConn = make(chan TReq)
	TimeMesg = make(chan TReq)
	go TimeCalc(timeConn, "CONNTIME")
	go TimeCalc(TimeMesg, "MESGTIME")

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

//AddDevice is used to add a device to the controller
func (c *Controller) AddDevice(raddr *net.TCPAddr) {
	c.NumDevices++
	d := NewDevice(*raddr)
	c.DBuffer <- d
}

//ListenToCascade Built off the exampe GOlang code
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
		timeConn <- TReq{*conn, time.Now()}
		if err != nil {
			panic(err)
		}
		msg, _ := ReadMessage(*conn)
		var msgType uint8 = uint8(msg[4])
		// If OPEN_CONN_1
		if msgType == 38 {
			// Instantiate a goroutine
			gid := atomic.AddUint32(&c.idCounter, 1)
			ch := make(chan CReq, chansize)
			c.chmap[gid] = ch
			go c.OperateDeviceOnInstance(gid, msg, ch, *conn)
		} else {
			msgGid := ReadInt32(msg[5:9])
			// If we are a OPEN_CONN_2
			c.chmap[msgGid] <- CReq{*conn, msg}
		}
	}
}

//OperateDeviceOnInstance We instantiate one of these per cascade instance
func (c *Controller) OperateDeviceOnInstance(gid uint32, initMsg []byte, ch chan CReq, oc1 net.TCPConn) {
	// Get device
	var err error
	dev := <-c.DBuffer
	conn := dev.GetOC1()
	//Timing
	timeConn <- TReq{oc1, time.Now()}
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
		//Timing
		timeConn <- TReq{newCReq.Conn, time.Now()}
		_, err = Handshake(newCReq.Conn, conn, newCReq.Buff, gid)
		if err == nil {
			go dev.DoForwarding(newCReq.Conn, conn)
		}
	}
}

//TimeCalc used to help calculate the timing issues between things
func TimeCalc(ch chan TReq, header string) {
	connMap := make(map[net.TCPConn]time.Time)
	for {
		blah := <-timeConn
		t, ok := connMap[blah.Conn]
		if !ok {
			connMap[blah.Conn] = blah.Time
		} else {
			elapsed := blah.Time.Sub(t)
			delete(connMap, blah.Conn)
			fmt.Printf("%s: %s\n", header, elapsed)
		}
	}
}
