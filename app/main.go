package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Client connected %s", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	var messageBuffer string

	for {
		messagePart, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client")
			} else {
				log.Println(err)
			}
			break
		}

		messageBuffer += messagePart

		messages := strings.Split(messageBuffer, "\r\n")
		if len(messages) > 1 {
			for _, message := range messages[:len(messages)-1] {
				if strings.HasPrefix(message, "!AIVDM") {
					log.Printf("Received message: %s\n", message)
				}
			}
			messageBuffer = messages[len(messages)-1]
		}
	}
}
