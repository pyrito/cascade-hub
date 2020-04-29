package main
import(
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type RPCMessage struct {
	msg_type uint8
	msg_pid uint32
	msg_eid uint32
	msg_n uint32
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
	full_msg := make([]byte, 0)
	buff := make([]byte, 256)
	bytes_read := 0

	// Temporary fix, should not rely on timeout
	err := n.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	if err != nil {
		fmt.Println("SetReadDeadline failed:", err)
	}
	// Read the amount of data needed
	for {
		// read the full message, or return an error
		num_read, err := n.Read(buff)
		// fmt.Printf("received %d\n", num_read)
		// fmt.Println(buff[:num_read])
		if err != nil {
			//panic(err)
			break
		}
		bytes_read += num_read
		full_msg = append(full_msg, buff[:num_read]...)
	}

	// fmt.Printf("======================\n")
	// type_ := uint8(full_msg[0])
	// pid_ := ReadInt32(full_msg[1:5])
	// eid_ := ReadInt32(full_msg[5:9])
	// n_ := ReadInt32(full_msg[9:13])

	fmt.Println(full_msg)
	return full_msg[:bytes_read]
}



