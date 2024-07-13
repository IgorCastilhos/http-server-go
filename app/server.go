package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var (
	maxConnections = 5
	semaphore      = make(chan struct{}, maxConnections)
	filesDirectory string
)

func handleConnection(conn net.Conn) {

	defer conn.Close()

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}

	request := string(buffer[:n])
	lines := strings.Split(request, "\r\n")
	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")
	method := parts[0]
	path := strings.Split(lines[0], " ")[1]

	userAgent := ""
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "User-Agent: ") {
			userAgent = line[len("User-Agent: "):]
			break
		}
	}

	gzipSupported := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Accept-Encoding:") && strings.Contains(line, "gzip") {
			gzipSupported = true
			break
		}
	}

	response := ""
	if strings.HasPrefix(path, "/files/") {
		filename := path[len("/files/"):]

		if method == "POST" {
			body := lines[len(lines)-1]

			filePath := filesDirectory + "/" + filename
			err := os.WriteFile(filePath, []byte(body), 0644)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))

		} else if method == "GET" {

			filePath := filesDirectory + "/" + filename // Combine the directory path with the filename
			fileContents, err := os.ReadFile(filePath)

			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}

			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(fileContents))))
			conn.Write(fileContents)
		}
	} else if path == "/user-agent" && userAgent != "" {

		responseBody := userAgent
		contentLength := len(responseBody)

		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, responseBody)
	} else if strings.HasPrefix(path, "/echo/") {
		echoStr := path[len("/echo/"):]

		var compressedData bytes.Buffer
		gz := gzip.NewWriter(&compressedData)
		_, err := gz.Write([]byte(echoStr))
		if err != nil {
			fmt.Println("Error writing to gzip writer: ", err)
			return
		}
		err = gz.Close()
		if err != nil {
			fmt.Println("Error closing gzip writer: ", err)
			return
		}

		if gzipSupported {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n", compressedData.Len())))
			io.Copy(conn, &compressedData)
		} else {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)))
		}

	} else if path == "/index.html" || path == "/" {
		if gzipSupported {
			response = "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\n\r\n"
		} else {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		}
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to connection: ", err)
	}
	<-semaphore
}

func main() {
	flag.StringVar(&filesDirectory, "directory", ".", "the directory to serve files from")
	flag.Parse()

	if filesDirectory == "" {
		fmt.Println("Please provide the directory to serve files from using the -directory flag")
		os.Exit(1)
	}
	fmt.Println("Logs from your program will appear here!")

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
		semaphore <- struct{}{}

		go handleConnection(conn)
	}
}
