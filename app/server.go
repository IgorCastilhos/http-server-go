package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the request
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	// Log the request (optional)
	fmt.Println("Received request:", strings.TrimSpace(request))

	// Set HTTP version and status line
	statusLine := "HTTP/1.1 200 OK\r\n"

	// Set headers
	headers := "Content-Type: text/plain\r\n"
	headers += "Custom-Header: CustomValue\r\n"

	// End headers with CRLF
	headers += "\r\n"

	// Set body
	body := "Hello, this is the body of the response.\r\n"

	// Write response to the connection
	response := statusLine + headers + body
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
		return
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	go handleConnection(conn)
}
