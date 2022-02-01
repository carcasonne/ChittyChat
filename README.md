# ChittyChat

This project was created for the course Distributed Systems at the IT University of Copenhagen. 

This is a chat service written in golang with a single server handling all the client requests, which can be run on your local machine.
The client and server communicate via gRPC implemented with protobuf.

## To run

Setup server. 
Enter the server directory first: 

`
go run .
`

Create client:
Enter the client directory:

`
go run .
`

Do this for as many clients you want in the chat

## Log

The server logs all messages to server/log.txt

## Lamport

Lamport code can be found in ChittyChat/LamportTime.go
