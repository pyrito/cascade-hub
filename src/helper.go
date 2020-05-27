package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"fmt"
	"math"
)

type RPCMessage struct {
	msgType uint8
	msgPid  uint32
	msgEid  uint32
	msgN    uint32
}

/* Credits to StackOverFlow */
// TODO: might be bug in using uint32...
func ReadUInt32(data []byte) uint32 {
	var ret uint32 = 0
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return ret
}

/* Read the message from the connection point */
func ReadMessage(n net.TCPConn) ([]byte, error) {
	fullMsg := make([]byte, 0)
	header := make([]byte, 4)
	numRead, err := n.Read(header)

	if err != nil {
		if err == io.EOF {
			return fullMsg, io.EOF
		}
		panic(err)
	}
	if numRead < 4 {
		panic("Protocol invariant Violated")
	}

	// TODO: find a better way of doing this, this is jank
	numToRead := int(ReadUInt32(header[0:4]))
	buff := make([]byte, numToRead)
	numToRead += 4
	totalRead := numRead
	fullMsg = append(fullMsg, header[:4]...)

	for {
		numToRead -= numRead
		// read the full message, or return an error
		if numToRead <= 0 {
			break
		}
		numRead, err = n.Read(buff)
		
		if err != nil {
			panic(err)
		}
		fullMsg = append(fullMsg, buff[:numRead]...)
		totalRead += numRead
	}
	//fmt.Println(fullMsg[:totalRead])
	return fullMsg[:totalRead], nil
}

func ReadFromConn(conn net.TCPConn, ch chan []byte) {
	for {
		res, err := ReadMessage(conn)

		if err == io.EOF {
			conn.Close()
			close(ch)
			return
		}
		fmt.Println("sending to chan")
		ch <- res
	}
}

func Handshake(conn0 net.TCPConn, conn1 net.TCPConn, msg []byte, gid uint32) (uint32, error) {
	_, err := conn1.Write(msg)
	if err != nil {
		panic(err)
	}
	res, err := ReadMessage(conn1)
	if err == io.EOF {
		conn0.Close()
		conn1.Close()
		return 0, err
	}
	pid := TranslateGIDPID(&res, gid)
	_, err = conn0.Write(res)
	if err != nil {
		panic(err)
	}
	return pid, nil
}

// TODO: Messy code, covering just edge case work
func TranslateGIDPID(msg *[]byte, toins uint32) uint32 {	
	old := uint32(0)
	off := uint32(0)
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, toins)

	len := ReadUInt32((*msg)[off:off+4])
	off += 4
	typ := uint8((*msg)[off])

	// In the case that we have more than one message packed in, they're all going the same way
	for {
		off += 1
		old = ReadUInt32((*msg)[off:off+4])
		copy((*msg)[off:off+4], temp[:])
		off += 12

		// If set_input or set_state
		if typ == 5 || typ == 7 {
			n := ReadUInt32((*msg)[off:off+4])
			off += 4
			for i := uint32(0); i < n; i++ {
				off += 4
				bits := ((1 << 30) - 1) & ReadUInt32((*msg)[off:off+4])
				off += uint32(math.Ceil(float64(bits) / 8)) + 4
			}
		} else if typ == 13 {
			off += 4
			bits := ((1 << 30) - 1) & ReadUInt32((*msg)[off:off+4])
			// Make sure to read the lower 30 bits
			off += uint32(math.Ceil(float64(bits) / 8)) + 4
		} else if typ == 2 { 
			// If a compilation message
			break
		}

		if off == (len + 4) {
			break
		}

		typ = uint8((*msg)[off])
	}

	return old
}
