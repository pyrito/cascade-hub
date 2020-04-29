package main

import "net"

/** The purpose of this struct is to encapsulate the FPGA client **/
type Device struct {
	Conn       *net.TCPConn
	JobRunning bool
	ErrorCode  int
}

func NewDevice(conn *net.TCPConn, running bool, err int) *Device {
	var d *Device = new(Device)
	d.Conn = conn
	d.JobRunning = running
	d.ErrorCode = err
	return d
}
