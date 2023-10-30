package main

import (
	"io"
	"log"
	"net"
)

func main() {
	log.Println("Server is running on port 2001")

	l, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	buf := make([]byte, 128)

	for {
		n, err := io.ReadFull(c, buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client")
			} else if err == io.ErrUnexpectedEOF {
				log.Println("Received less data than expected")
			} else {
				log.Println(err)
			}
			break
		}

		log.Printf("Received message: %s\n", string(buf[:n]))
	}

	defer c.Close()
}
