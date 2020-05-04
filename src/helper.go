package main

import (
	"bytes"
	"encoding/binary"
	"net"
)

type RPCMessage struct {
	msgType uint8
	msgPid  uint32
	msgEid  uint32
	msgN    uint32
}

/* Credits to StackOverFlow */
// TODO: might be bug in using uint32...
func ReadInt32(data []byte) uint32 {
	var ret uint32 = 0
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return ret
}

/* Read the message from the connection point */
func ReadMessage(n net.Conn) []byte {
	fullMsg := make([]byte, 0)
	buff := make([]byte, 256)
	numRead, err := n.Read(buff)
	if numRead < 4 {
		panic("Protocol invariant Violated")
	}

	numToRead := ReadInt32(buff[0:3])
	numToRead += 4
	totalRead := numRead

	for {
		fullMsg = append(fullMsg, buff[:numRead])
		numToRead -= numRead
		// read the full message, or return an error
		if numToRead <= 0 {
			break
		}
		numRead, err := n.Read(buff)
		if err != nil {
			panic(err)
		}
		totalRead += numRead
	}

	return fullMsg[:totalRead]
}
