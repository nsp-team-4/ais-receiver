package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"ais-receiver/ais"
)

func RunServer() {
	listener := createListener()
	defer listener.Close()

	log.Println("Server is running on port 2001")

	for {
		acceptAndHandleConnection(listener)
	}
}

func createListener() net.Listener {
	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Println(err)
	}

	return listener
}

func acceptAndHandleConnection(listener net.Listener) {
	err := acceptConnection(listener)
	if err != nil {
		log.Println(err)
	}
}

func acceptConnection(listener net.Listener) error {
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %s", err)
	}

	go handleClient(conn)

	return nil
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	log.Printf("Client connected %s", conn.RemoteAddr().String())
	handleClientConnection(conn)
}

func handleClientConnection(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	processMessages(scanner)
}

func processMessages(scanner *bufio.Scanner) {
	for scanner.Scan() {
		processMessage(scanner)
	}

	handleScannerError(scanner)
}

func handleScannerError(scanner *bufio.Scanner) {
	err := scanner.Err()
	if err != nil {
		log.Println("Error reading:", err)
	}
}

func processMessage(scanner *bufio.Scanner) {
	message := scanner.Text()
	handleMessage(message)
}

func handleMessage(message string) {
	log.Println(message)
	err := ais.HandleMessage(message)
	if err != nil {
		log.Println(err)
	}
}
