package main
import (
	"fmt"
	"golang.org/x/crypto/ssh"
	//"io/ioutil"
	// "io"
	// "os"
)

/** One type of controller scheduling property **/
/** TODO: create some kind of interface that accepts different schedulers **/
type Controller struct {
	ID int
	NumDevices int
	DevicesAvailable map[int]*Device
	Buffer DeviceBuffer
}

/* Ideally this needs to request the devices from another interface */
func (c *Controller) Initialize(devices int) {
	// TODO: give proper ID
	c.ID = 0
	c.NumDevices = devices
	c.DevicesAvailable = make(map[int]*Device)
	c.Buffer.Initialize(devices)
	// Add to the map and add to the buffer as well...
	for i := 0; i < devices; i++ {
		// Create some dummy device
		// Ideally we want to provision some n devices from switch, etc.
		d := NewDevice("192.168.1.1", i, false, false, -1)
		c.Add(i, d)
		c.Buffer.Enqueue(d)
	}
}

/* Add a device to the controller */
func (c *Controller) Add(id int, d *Device) {
	devices := c.DevicesAvailable
	if _, ok := devices[id]; ok {
		fmt.Println("Rewriting existing value")
	} else {
		devices[id] = d
	}
}

/* Remove a device from the controller */
func (c *Controller) Remove(id int) {
	available := c.DevicesAvailable
	if device, ok := available[id]; ok {
		device.Removed = true
		delete(available, id)
	}
}

/* Send the work to some provisioned FPGA */
func (c *Controller) Execute(filePath string) int {
	var d *Device = c.Buffer.Dequeue()
	// Add the SSH logic in here accordingly
	// Think about security or something probably
	fmt.Printf("d.id: %d\n", d.ID)
	_, err := ConnectToDevice(d.IPAddress)
	if err != nil {
		panic(err)
	}

	return 0
}

/* Special thanks to https://medium.com/tarkalabs/ssh-recipes-in-go-part-one-5f5a44417282 */
func ConnectToDevice(hostname string) (*ssh.Session, error) {
	// key, err := ioutil.ReadFile("/u/vkarthik/.ssh/id_rsa")
	// if err != nil {
	// 	panic(err)
	// }

	// signer, err := ssh.ParsePrivateKey(key)
	// if err != nil {
	// 	panic(err)
	// }
	// Temporary placeholder, log into the machines
	config := &ssh.ClientConfig {
		User: "vkarthik",
		Auth: []ssh.AuthMethod{ 
			ssh.Password(""),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	fmt.Println("created client config")
	conn, err := ssh.Dial("tcp", "128.83.144.171:22", config)

	if err != nil {
		return nil, err
	}

	sess, err := conn.NewSession()
	defer sess.Close()

	if err != nil {
		return nil, err
	}
	
	err = sess.Run("ls") // eg., /usr/bin/whoami
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func SCPToDevice() {
	// Need to copy over 
}
// func ShowSSHOutput() {
// 	// Show the std out from the SSH connection
// 	sessStdOut, err := sess.StdoutPipe()
// 	if err != nil {
// 		panic(err)
// 	}
// 	go io.Copy(os.Stdout, sessStdOut)

// 	// Show the std error from the SSH connection
// 	sessStderr, err := sess.StderrPipe()
// 	if err != nil {
// 		panic(err)
// 	}
// 	go io.Copy(os.Stderr, sessStderr)
// }