package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

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
	header := make([]byte, 8)
	numRead, err := n.Read(header)

	if err != nil {
		if err == io.EOF {
			return fullMsg, io.EOF
		}
		panic(err)
	}
	if numRead < 8 {
		panic("Protocol invariant Violated")
	}

	// TODO: find a better way of doing this, this is jank
	numToRead := int(ReadUInt32(header[0:8]))
	buff := make([]byte, numToRead)
	numToRead += 8
	totalRead := numRead
	fullMsg = append(fullMsg, header[:8]...)

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
		ch <- res
	}
}

func Handshake(conn0 net.TCPConn, conn1 net.TCPConn, msg []byte, gid uint32) error {
	InsertGID(&msg, gid)
	_, err := conn1.Write(msg)
	if err != nil {
		panic(err)
	}
	res, err := ReadMessage(conn1)
	if err == io.EOF {
		conn0.Close()
		conn1.Close()
		return err
	}
	_, err = conn0.Write(res)
	if err != nil {
		panic(err)
	}
	return nil
}

func InsertGID(msg *[]byte, toins uint32) {
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, toins)
	copy((*msg)[4:8], temp[:])
}
