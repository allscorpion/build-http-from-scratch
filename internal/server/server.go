package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/allscorpion/build-http-from-scratch/internal/request"
	"github.com/allscorpion/build-http-from-scratch/internal/response"
)

type Server struct {
	Listener net.Listener
	isOpen   atomic.Bool
	Handler  Handler
}

type HandlerError struct {
	StatusCode   response.StatusCode
	ErrorMessage string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := Server{
		Listener: listener,
		isOpen:   atomic.Bool{},
		Handler:  handler,
	}

	server.isOpen.Store(true)

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	defer func() {
		fmt.Println("The connection has been closed")
	}()
	s.isOpen.Store(false)
	return s.Listener.Close()
}

func (s *Server) listen() {
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
	req, err := request.RequestFromReader(conn)
	responseWriter := response.NewWriter(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.InternalServerErrorStatus)
		body := err.Error()
		responseWriter.WriteHeaders(response.GetDefaultHeaders(len(body)))
		responseWriter.WriteBody(body)
		return
	}

	s.Handler(responseWriter, req)
}
