package main

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/vishalkuo/bimap"
)

type PIDPort struct {
	PID  uint64
	Port uint64
}

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
	chmap      map[uint64]CReq
	idCounter  uint64
}

var chansize = 1000

/* Ideally this needs to request the devices from another interface */
/** TODO need a connection finding phase **/
func (c *Controller) Initialize(devices int) {
	c.NumDevices = 0
	c.DBuffer = make(chan *Device, 1000)
	c.CReqChan = make(chan CReq, 1000)
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
		cascades[conn] = make(chan []byte, 1000)
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
		var msgType uint8 = uint8(msg[4])
		// If OPEN_CONN_1
		if msgType == 37 {
			// Instantiate a goroutine
			gid := atomic.AddU64(&c.idCounter, 1)
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

func readFromConn(conn net.TCPConn, ch chan []byte) {
	res := ReadMessage(conn)
	ch <- res

}

func handshake(conn0 net.TCPConn, conn1 net.TCPConn, msg []byte) {
	_, err := conn1.Write(msg)
	if err != nil {
		panic(err)
	}
	res := ReadMessage(conn1)
	_, err = conn0.Write(res)
	if err != nil {
		panic(err)
	}
}

/** We instantiate one of these per cascade instance **/
func (c *Controller) OperateDeviceOnInstance(gid uint64, initMsg []byte, ch chan CReq, oc1 net.TCPConn) {
	// Get device
	dev := <-c.DBuffer
	conn := dev.getOC1()
	handshake(oc1, conn)
	go doForwarding(oc1, conn)

	for {
		newCReq := <-ch
		msgType := uint8(newCreq.msg[4])
		if msgType == 38 {
			conn = dev.getOC2()
		} else {
			conn = dev.getNextOpenConn()
		}
		handshake(newCReq.Conn, conn)
		go doForwarding(newCreq.Conn, conn)
	}
}
