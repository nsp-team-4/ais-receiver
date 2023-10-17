package main

import (
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
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Received data:", string(buf[:n]))
	}
}
