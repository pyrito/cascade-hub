package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
func ReadMessage(n net.TCPConn) []byte {
	fmt.Println(n.RemoteAddr().String())
	fullMsg := make([]byte, 0)
	buff := make([]byte, 256)
	numRead, err := n.Read(buff)
	fmt.Println("Passed a read")
	if err != nil {
		panic(err)
	}
	if numRead < 4 {
		panic("Protocol invariant Violated")
	}

	numToRead := ReadInt32(buff[0:4])
	numToRead += 4
	totalRead := numRead

	for {
		fullMsg = append(fullMsg, buff[:numRead]...)
		numToRead -= uint32(numRead)
		// read the full message, or return an error
		if numToRead <= 0 {
			break
		}
		numRead, err = n.Read(buff)
		fmt.Println("Passed a read")
		if err != nil {
			panic(err)
		}
		totalRead += numRead
	}

	return fullMsg[:totalRead]
}

func ReadFromConn(conn net.TCPConn, ch chan []byte) {
	res := ReadMessage(conn)
	ch <- res
}

func Handshake(conn0 net.TCPConn, conn1 net.TCPConn, msg []byte, gid uint32) uint32 {
	_, err := conn1.Write(msg)
	if err != nil {
		panic(err)
	}
	res := ReadMessage(conn1)
	pid := TranslateGIDPID(&res, gid)
	_, err = conn0.Write(res)
	if err != nil {
		panic(err)
	}
	return pid
}

func TranslateGIDPID(msg *[]byte, toins uint32) uint32 {
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, toins)
	old := ReadInt32((*msg)[5:9])
	copy((*msg)[5:9], temp[:])
	return old
}
