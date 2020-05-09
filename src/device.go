package main

import (
	"net"
	"sync"
	"time"
)

//List of outbound ports
var outBoundP []net.TCPAddr
var readOutP sync.RWMutex

// The Device encapsulates the FPGA client
type Device struct {
	GID       uint32
	PID       uint32
	connIndex int
	conns     []net.TCPConn
	raddr     net.TCPAddr
	lock      sync.RWMutex
}

func InitializeDeviceManagement() {
	outBoundP = make([]net.TCPAddr, 2)
	addr0, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		panic(err)
	}
	outBoundP[0] = *addr0
	addr1, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		panic(err)
	}
	outBoundP[1] = *addr1
}

func NewDevice(raddr net.TCPAddr) *Device {
	var d *Device = new(Device)
	d.conns = make([]net.TCPConn, 2)
	d.raddr = raddr
	readOutP.RLock()
	conn0, err := net.DialTCP("tcp", &outBoundP[0], &raddr)
	if err != nil {
		panic(err)
	}
	conn0.SetKeepAlive(true)
	d.conns[0] = *conn0
	conn1, err := net.DialTCP("tcp", &outBoundP[1], &raddr)
	if err != nil {
		panic(err)
	}
	conn1.SetKeepAlive(true)
	d.conns[1] = *conn1
	readOutP.RUnlock()
	d.connIndex = 2
	return d
}

func (d *Device) DoForwarding(connCI net.TCPConn, connCD net.TCPConn) {
	chCI := make(chan []byte)
	go ReadFromConn(connCI, chCI)
	chCD := make(chan []byte)
	go ReadFromConn(connCD, chCD)

	for {
		select {
		case msg, ok := <-chCI:
			if !ok {
				return
			}
			gid := ReadInt32(msg[5:9])
			if gid != d.GID {
				panic("YOU HAVE A DIFFERENT GID")
			}
			TranslateGIDPID(&msg, d.PID)
			//Timing
			TimeMesg <- TReq{connCI, time.Now()}
			_, err := connCD.Write(msg)
			if err != nil {
				panic(err)
			}
		case msg, ok := <-chCD:
			if !ok {
				return
			}
			pid := ReadInt32(msg[5:9])
			if pid != d.PID {
				panic("YOU HAVE A DIFFERENT PID")
			}
			TranslateGIDPID(&msg, d.GID)
			//Timing
			TimeMesg <- TReq{connCI, time.Now()}
			_, err := connCI.Write(msg)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (d *Device) GetOC1() net.TCPConn {
	return d.conns[0]
}

func (d *Device) GetOC2() net.TCPConn {
	return d.conns[1]
}

func (d *Device) GetNextConn() net.TCPConn {
	if d.connIndex >= len(d.conns) {
		//Realize we need to access the global structure
		readOutP.RLock()
		var finalLAddr net.TCPAddr
		if d.connIndex >= len(outBoundP) {
			readOutP.RUnlock()
			readOutP.Lock()
			if d.connIndex >= len(outBoundP) {
				laddr, err := net.ResolveTCPAddr("tcp", ":0")
				if err != nil {
					panic(err)
				}
				outBoundP = append(outBoundP, *laddr)
			}
			finalLAddr = outBoundP[d.connIndex]
			readOutP.Unlock()
		} else {
			finalLAddr = outBoundP[d.connIndex]
			readOutP.RUnlock()
		}
		rconn, err := net.DialTCP("tcp", &finalLAddr, &d.raddr)
		if err != nil {
			panic(err)
		}
		d.conns = append(d.conns, *rconn)
	}
	conn := d.conns[d.connIndex]
	d.connIndex++
	return conn
}
