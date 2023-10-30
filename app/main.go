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
	// defer c.Close()

	buf := make([]byte, 1024)

	for {
		n, err := io.ReadFull(c, buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client")
				return
			} else if err == io.ErrUnexpectedEOF {
				log.Println("Received less data than expected")
			} else {
				log.Println(err)
			}
			break
		}

		log.Printf("Received message: %s\n", string(buf[:n]))
	}
}

// import (
// 	"fmt"
// 	"log"
// 	"net"
// )

// type Message struct {
// 	from    string
// 	payload []byte
// }

// type Server struct {
// 	listenAddr string
// 	ln         net.Listener
// 	quitch     chan struct{}
// 	msgch      chan Message
// }

// func NewServer(listenAddr string) *Server {
// 	return &Server{
// 		listenAddr: listenAddr,
// 		quitch:     make(chan struct{}),
// 		msgch:      make(chan Message, 10),
// 	}
// }

// func (s *Server) Start() error {
// 	ln, err := net.Listen("tcp", s.listenAddr)
// 	if err != nil {
// 		return err
// 	}
// 	defer ln.Close()
// 	s.ln = ln
// 	go s.acceptLoop()

// 	<-s.quitch
// 	close(s.msgch)

// 	return nil
// }

// func (s *Server) acceptLoop() {
// 	for {
// 		conn, err := s.ln.Accept()
// 		if err != nil {
// 			fmt.Println("accept error:", err)
// 			continue
// 		}

// 		fmt.Println("new connection to the server:", conn.RemoteAddr())

// 		go s.readLoop(conn)
// 	}
// }

// func (s *Server) readLoop(conn net.Conn) {
// 	defer conn.Close()
// 	buf := make([]byte, 2048)

// 	for {
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			fmt.Println("read error", err)
// 			continue
// 		}

// 		s.msgch <- Message{
// 			from:    conn.RemoteAddr().String(),
// 			payload: buf[:n],
// 		}

// 		conn.Write([]byte("ok"))
// 	}
// }

// func main() {
// 	server := NewServer(":2001")

// 	go func() {
// 		for msg := range server.msgch {
// 			str := fmt.Sprintf("received message from connection (%s):%s\n", msg.from, string(msg.payload))
// 			fmt.Println(msg.payload)
// 			fmt.Println(str)
// 		}
// 	}()

// 	log.Fatal(server.Start())
// }
