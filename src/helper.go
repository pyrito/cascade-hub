package main
import(
	"bytes"
	"encoding/binary"
	"io"
	"bufio"
	"fmt"
	"net"
)

type RPCMessage struct {
	msg_type uint8
	msg_pid uint32
	msg_eid uint32
	msg_n uint32
}

/* Convert the message struct to bytes */
func MsgToBytes(msg RPCMessage) []byte {
	buff := make([]byte, 13)
	buff[0] = byte(msg.msg_type)

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, msg.msg_pid)
	for j := 0; j < len(b); j++ {
		buff[1 + j] = b[j]
	}

	b1 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b1, msg.msg_eid)
	for j := 0; j < len(b1); j++ {
		buff[5 + j] = b1[j]
	} 

	b2 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b2, msg.msg_eid)
	for j := 0; j < len(b2); j++ {
		buff[9 + j] = b2[j]
	}
	fmt.Println(len(buff))
	return buff
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
func ReadMessage(n net.Conn) RPCMessage {
	buff := make([]byte, 17)
	reader := bufio.NewReader(n)
	retry := true
	// Read the amount of data needed
	for retry {
		// read the full message, or return an error
		num_read, err := io.ReadFull(reader, buff)
		fmt.Printf("received %d\n", num_read)
		if err == nil {
			retry = false
		}
		fmt.Printf("waiting\n")
	}

	type_ := uint8(buff[4])
	pid_ := ReadInt32(buff[5:9])
	eid_ := ReadInt32(buff[9:13])
	n_ := ReadInt32(buff[13:17])

	fmt.Printf("type: %d, pid: %d, eid: %d, n: %d\n", type_, pid_, eid_, n_)
	return RPCMessage{type_, pid_, eid_, n_}
}



