package main
import (
	"flag"
	"fmt"
	"net/http"
)

var controller Controller

func handleJobReq(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Received some job request...\n")
	//controller.Execute("somepath")
	fmt.Fprintf(w, "Completed job request...\n")
}

func main() {
	// input := flag.String("input", "", "string-valued path to the running program")
	devices := flag.Int("devices", 0, "number of devices to set up")
	
	flag.Parse()
	http.HandleFunc("/req", handleJobReq)

	fmt.Printf("workers: %d\n", *devices)
	
	controller.Initialize(*devices)
	fmt.Printf("controller devices: %d\n", controller.NumDevices)
	//http.ListenAndServe(":8090", nil)
	controller.ListenToCascade()
}