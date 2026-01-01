package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

func WriteHandleError(w io.Writer, handleError *HandlerError) error {
	err := response.WriteStatusLine(w, response.StatusCode(handleError.StatusCode))

	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(handleError.ErrorMessage))
	err = response.WriteHeaders(w, headers)

	if err != nil {
		return err
	}

	err = response.WriteBody(w, handleError.ErrorMessage)

	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		handlerError := HandlerError{
			StatusCode:   response.InternalServerErrorStatus,
			ErrorMessage: err.Error(),
		}
		WriteHandleError(conn, &handlerError)
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0))
	handlerError := s.Handler(buffer, req)

	if handlerError != nil {
		WriteHandleError(conn, handlerError)
		return
	}

	responseBody := buffer.String()

	defaultHeaders := response.GetDefaultHeaders(len(responseBody))
	response.WriteStatusLine(conn, response.OKStatus)
	response.WriteHeaders(conn, defaultHeaders)
	response.WriteBody(conn, responseBody)
}
