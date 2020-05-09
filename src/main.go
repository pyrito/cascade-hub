package main

import (
	"flag"
	"fmt"
)

var controller Controller

func main() {
	// input := flag.String("input", "", "string-valued path to the running program")
	devices := flag.Int("devices", 0, "number of devices to set up")

	flag.Parse()

	if *devices == 0 {
		panic("You have created no devices")
	}

	fmt.Printf("workers: %d\n", *devices)

	controller.Initialize(*devices)
	fmt.Printf("controller devices: %d\n", controller.NumDevices)
	//http.ListenAndServe(":8090", nil)
	controller.ListenToCascade()
}
