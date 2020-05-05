package main

import (
	"net"
	"sync"
)

//List of outbound ports
var outBoundP []net.TCPAddr
var readOutP sync.RWMutex

/** The purpose of this struct is to encapsulate the FPGA client **/
type Device struct {
	GID       uint32
	PID       uint32
	connIndex int
	conns     []net.TCPConn
	raddr     net.TCPAddr
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
	d.conns[0] = *conn0
	conn1, err := net.DialTCP("tcp", &outBoundP[1], &raddr)
	if err != nil {
		panic(err)
	}
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
		case msg := <-chCI:
			gid := ReadInt32(msg[5:9])
			if gid != d.GID {
				panic("YOU HAVE A DIFFERENT GID")
			}
			TranslateGIDPID(&msg, d.PID)
			_, err := connCD.Write(msg)
			if err != nil {
				panic(err)
			}
		case msg := <-chCD:
			pid := ReadInt32(msg[5:9])
			if pid != d.PID {
				panic("YOU HAVE A DIFFERENT PID")
			}
			TranslateGIDPID(&msg, d.GID)
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
		readOutP.RLock()
		if d.connIndex >= len(outBoundP) {
			readOutP.RUnlock()
			readOutP.Lock()
			if d.connIndex >= len(outBoundP) {
				laddr, err := net.ResolveTCPAddr("tcp", ":0")
				if err != nil {
					panic(err)
				}
				outBoundP = append(outBoundP, *laddr)
				rconn, err := net.DialTCP("tcp", &outBoundP[d.connIndex], &d.raddr)
				if err != nil {
					panic(err)
				}
				d.conns = append(d.conns, *rconn)
			}
			readOutP.Unlock()
		} else {
			rconn, err := net.DialTCP("tcp", &outBoundP[d.connIndex], &d.raddr)
			if err != nil {
				panic(err)
			}
			d.conns[d.connIndex] = *rconn
			readOutP.RUnlock()
		}
	}
	conn := d.conns[d.connIndex]
	d.connIndex++
	return conn
}
