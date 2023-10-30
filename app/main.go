package main

import (
	"fmt"
	"net"
)

func main() {
	// Define the address and port for the server
	address := "0.0.0.0" // Bind to all available network interfaces
	port := "2001"       // Change this to your desired port number

	// Create a TCP listener
	listener, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("AIS server is listening on %s:%s\n", address, port)

	for {
		// Accept incoming connections and start a goroutine to handle each connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Initialize a buffer to read data from the connection
	buffer := make([]byte, 1024)

	for {
		// Read data from the connection into the buffer
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}

		// Process the received data (AIS messages) here
		data := buffer[:n]
		aisMessage := string(data)

		fmt.Printf("Received AIS message: %s\n", aisMessage)
	}
}
