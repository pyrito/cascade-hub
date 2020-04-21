package main
// import (

// )

/** The purpose of this struct is to encapsulate the FPGA client **/
type Device struct {
	IPAddress string
	ID int
	JobRunning bool
	Removed bool
	ErrorCode int
}

func NewDevice(addr string, id int, running bool, removed bool, err int) *Device {
	var d *Device = new(Device)
	d.IPAddress = addr
	d.ID = id
	d.JobRunning = running
	d.Removed = removed
	d.ErrorCode = err
	return d
}

func EmptyDevice() *Device {
	var d *Device = new(Device)
	d.IPAddress = ""
	d.ID = -1
	d.JobRunning = false
	d.Removed = false
	d.ErrorCode = -1
	return d
}
