package server

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	isOpen   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := Server{
		Listener: listener,
		isOpen:   atomic.Bool{},
	}

	server.isOpen.Store(true)

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.isOpen.Store(false)
	return s.Listener.Close()
}

func (s *Server) listen() {
	defer func() {
		s.Close()
		fmt.Println("The connection has been closed")
	}()

	for {
		conn, err := s.Listener.Accept()

		if !s.isOpen.Load() {
			break
		}

		if err != nil {
			fmt.Printf("an error has occured accepted the connection %v\n", err)
			continue
		}

		fmt.Println("the connection has been accepted")

		go func() {
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	lines := []string{"HTTP/1.1 200 OK", "Content-Type: text/plain", "Content-Length: 13", "", "Hello World!"}
	conn.Write([]byte(strings.Join(lines, "\r\n")))
}
