package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

var (
	// Limit the number of concurrent connections
	maxConnections = 5
	semaphore      = make(chan struct{}, maxConnections)
	filesDirectory string
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

	// Initialize the response variable
	response := ""
	userAgent := ""

	// Loop through headers to find user-agent
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "User-Agent: ") {
			userAgent = line[len("User-Agent: "):]
			break
		}
	}

	// Check the path
	if strings.HasPrefix(path, "/files/") {
		filename := path[len("/files/"):]

		filePath := filesDirectory + "/" + filename // Combine the directory path with the filename
		fileContents, err := ioutil.ReadFile(filePath)

		if err != nil {
			// File does not exist
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}

		// File exists, serve it
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(fileContents))))
		conn.Write(fileContents)
	} else if path == "/user-agent" && userAgent != "" {

		// Prepare the response body
		responseBody := userAgent
		// Calculate the Content-Length
		contentLength := len(responseBody)

		// Construct the response with Content-Type and Content-Length headers
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, responseBody)
	} else if strings.HasPrefix(path, "/echo/") {
		// Extract the string after /echo/
		echoStr := path[len("/echo/"):]

		// Prepare the response body
		responseBody := echoStr
		// Calculate the Content-Length
		contentLength := len(responseBody)

		// Construct the response with Content-Type and Content-Length headers
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, responseBody)
	} else if path == "/index.html" || path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// Write the response to the connection
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err)
	}
	<-semaphore
}

func main() {
	// Parse the --directory flag
	flag.StringVar(&filesDirectory, "directory", ".", "the directory to serve files from")
	flag.Parse()

	// Ensure the directory flag is set
	if filesDirectory == "" {
		fmt.Println("Please provide the directory to serve files from using the -directory flag")
		os.Exit(1)
	}
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
		// Acquire a semaphore before handling the connection
		semaphore <- struct{}{}

		go handleConnection(conn)
	}
}
