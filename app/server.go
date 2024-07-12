package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Set the status code
	w.WriteHeader(http.StatusOK) // 200 OK

	// Add headers
	w.Header().Set("Content-Type", "text/plain")

	// Write Body
	_, err := fmt.Fprintln(w, "Hello, World!")
	if err != nil {
		return
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	http.HandleFunc("/", handler) // Handle requests to the root URL
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	_, err = l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
}
