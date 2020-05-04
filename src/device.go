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
	GlobalPID  uint32
	LocalPID   uint32
	connsIndex int
	conns      []net.TCPConn
	raddr      net.TCPAddr
}

func InitializeDeviceManagement() {
	outBoundP = make([]net.TCPAddr, 2)
	outBoundP[0] = net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	outBoundP[1] = net.ResolveTCPAddr("tcp", "127.0.0.1:0")
}

func NewDevice(raddr net.TCPAddr) *Device {
	var d *Device = new(Device)
	d.conns = make([]net.TCPConn, 2)
	d.raddr = raddr
	readOutP.RLock()
	d.conns[0] = net.DialTCP("tcp", outBoundP[0], raddr)
	d.conns[1] = net.DialTCP("tcp", outBoundP[1], raddr)
	readOutP.RUnlock()
	connsIndex = 2
	return d
}

func (d *Device) DoFowarding(connCI net.TCPConn, connCD net.TCPConn) {
	chCI := make(chan []byte)
	go readFromConn(connCI, chCI)
	chCD := make(chan []byte)
	go readFromConn(connCD, chCD)
	for {
		select {
		case msg <- chCI:
			_, err = connCD.Write(msg)
			if err != nil {
				panic(err)
			}
		case msg <- chCD:
			_, err = connCI.Write(msg)
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

func (d *Device) GetNextCon() net.TCPConn {
	if d.connIndex >= len(d.conns) {
		readOutP.RLock()
		if d.connIndex >= len(outBoundP) {
			readOutP.RUnlock()
			readOutP.Lock()
			if d.connIndex >= len(outBoundP) {
				outBoundP = append(outBoundP, net.ResolveTCPAddr(tcp, "127.0.0.1:0"))
				d.conns[d.connInndex] = net.DialTCP("tcp", outBoundP[d.connIndex], raddr)
			}
			readOutP.Unlock()
		} else {
			d.conns[d.connInndex] = net.DialTCP("tcp", outBoundP[d.connIndex], raddr)
			readOutP.RUnlock()
		}
	}
	conn := d.conns[d.connIndex]
	d.connIndex++
	return conn
}
