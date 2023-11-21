package tcp

import (
	"bufio"
	"log"
	"net"

	"ais-receiver/ais"
)

func RunServer() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Println(err)
	}

	defer listener.Close()

	for {
		acceptConnection(listener)
	}
}

func acceptConnection(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Client connected %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		err := ais.HandleMessage(message)
		if err != nil {
			log.Println(err)
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Println("Error reading:", err)
	}
}
