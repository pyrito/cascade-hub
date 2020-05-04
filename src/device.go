package main

import (
	"net"

	"github.com/vishalkuo/bimap"
)

/** The purpose of this struct is to encapsulate the FPGA client **/
type Device struct {
	Conn           *net.TCPConn
	JobRunning     bool
	ErrorCode      int
	GlobalLocalmap *bimap.BiMap
}

func NewDevice(conn *net.TCPConn, running bool, err int) *Device {
	var d *Device = new(Device)
	d.Conn = conn
	d.JobRunning = running
	d.ErrorCode = err
	d.GlobalLocalmap = bimap.NewBiMap()
	return d
}
