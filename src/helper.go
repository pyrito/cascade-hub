package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"
	"fmt"
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
func ReadMessage(n net.TCPConn) ([]byte, error) {
	fmt.Printf("Going to read from: ")
	fmt.Println(n.RemoteAddr().String())
	fullMsg := make([]byte, 0)
	header := make([]byte, 4)
	numRead, err := n.Read(header)
	fmt.Printf("Read from: ")
	fmt.Println(n.RemoteAddr().String())
	fmt.Println(header)
	if err != nil {
		if err == io.EOF {
			return fullMsg, io.EOF
		}
		panic(err)
	}
	if numRead < 4 {
		panic("Protocol invariant Violated")
	}

	numToRead := int(ReadInt32(header[0:4]))
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
	fmt.Println(fullMsg[:totalRead])
	return fullMsg[:totalRead], nil
}

func ReadFromConn(conn net.TCPConn, ch chan []byte) {
	for {
		res, err := ReadMessage(conn)
		fmt.Println("we are here after ReadMessage in ReadFromConn")
		fmt.Println(res)
		fmt.Println(err)
		//Timing
		TimeMesg <- TReq{conn, time.Now()}

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

func TranslateGIDPID(msg *[]byte, toins uint32) uint32 {
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, toins)
	old := ReadInt32((*msg)[5:9])
	copy((*msg)[5:9], temp[:])
	return old
}
