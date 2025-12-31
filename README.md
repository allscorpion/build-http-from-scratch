# Build HTTP from scratch

A project to learn how to build http from scratch. 

## Commands
Use the following commands to test the tcp listener

- go run ./cmd/tcplistener | tee {outputFileName}
- curl http://localhost:42069/

To test the udp listener you can use these commands 
- go run ./cmd/udpsender
- nc -u -l 42069