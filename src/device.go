package main

import (
	"net"

	"github.com/vishalkuo/bimap"
)

/** The purpose of this struct is to encapsulate the FPGA client **/
type Device struct {
	Conns          []net.TCPConn
	JobRunning     bool
	ErrorCode      int
	GlobalPID 	   uint32
	LocalPID	   uint32
	ConnsIndex	   int
}

func NewDevice(conn net.TCPConn, running bool, err int) *Device {
	var d *Device = new(Device)
	d.Conns = make([]net.TCPConn, 3)
	d.Conns[0] = conn
	d.JobRunning = running
	d.ErrorCode = err
	d.GlobalLocalmap = bimap.NewBiMap()
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

func (d *Device) setOC1(conn net.TCPConn) {
	d.Conns[1] = conn
}

func (d *Device) setOC2(conn net.TCPConn) {
	d.Conns[2] = conn
}

func (d *Device) getOC1() net.TCPConn {
	return d.Conns[1]
}

func (d *Device) getOC2() net.TCPConn{
	return d.Conns[2]
}

