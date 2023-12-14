package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"ais-receiver/ais"
)

func RunServer() {
	port, err := getPort()
	if err != nil {
		log.Fatal(err)
	}

	listener := createListener(port)
	defer listener.Close()

	log.Printf("Server is running on port %d\n", port)

	for {
		acceptAndHandleConnection(listener)
	}
}

func getPort() (int, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return strconv.Atoi(port)
}

func createListener(port int) net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
	err := ais.MessageReceiver(message)
	if err != nil {
		log.Println(err)
	}
}
