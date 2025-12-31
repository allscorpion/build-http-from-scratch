package main

import (
	"fmt"
	"net"

	"github.com/allscorpion/build-http-from-scratch/internal/request"
)

func main() {
	network, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Println("Error starting TCP connection:", err)
		return
	}

	defer func() {
		network.Close()
		fmt.Println("The connection has been closed")
	}()

	for {
		conn, err := network.Accept()

		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			return
		}

		fmt.Println("The connection has been accepted")
		req, err := request.RequestFromReader(conn)

		if err != nil {
			fmt.Printf("An error has occured getting the request %v", err)
			return
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", req.RequestLine.Method)
		fmt.Printf("- Target: %v\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %v: %v\n", key, value)
		}
		if len(req.Body) > 0 {
			fmt.Println("Body:")
			fmt.Printf("%v\n", string(req.Body))
		}
	}
}
