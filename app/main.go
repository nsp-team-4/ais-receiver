package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Starting server...")
	ln, err := net.Listen("tcp", ":2001")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	fmt.Println("Listening on port 2001...")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("New connection accepted.")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return
		}
		fmt.Printf("Message incoming: %s", string(message))
		conn.Write([]byte("Message received!\n"))
	}
}
