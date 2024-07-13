package main

/*
	This code extracts the URL path from an HTTP request,
	and responds with either a 200 or 404.
	1. Read the incoming data from the connection to get the
	HTTP request as a string.
	2. Split the request string by CRLF (\r\n) to separate
	the request line, headers, and body.
	3. Extract the request line (the first element of the
	split result) and then split it by spaces to get the
	method, path, and HTTP version.
	4. Check the path to determine the response. If the path
	matches a known route (e.g., /index.html), respond
	with 200 OK. Otherwise, respond with 404 Not Found.
	5. Construct the appropriate HTTP response based on
	the path check.
	6. Write the response back to the connection.
*/

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the request
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}
	request := string(buffer[:n])

	// Split request into lines
	lines := strings.Split(request, "\r\n")
	// Extract the request line
	requestLine := lines[0]
	// Split the request line into components
	parts := strings.Split(requestLine, " ")
	// Extract the path
	path := parts[1]

	// Determine the response based on the path
	response := ""
	if path == "/index.html" || path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Write the response to the connection
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err)
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
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}
