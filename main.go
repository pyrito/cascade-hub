package main
import (
	"flag"
	"fmt"
)

func main() {
	// input := flag.String("input", "", "string-valued path to the running program")
	devices := flag.Int("devices", 0, "number of devices to set up")
	
	flag.Parse()

	fmt.Printf("workers: %d\n", *devices)
	var controller Controller
	controller.Initialize(*devices)
	fmt.Printf("controller devices: %d\n", controller.NumDevices)
	controller.Execute("/path/to/somewhere")
}