package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	t "time"

	"github.com/Restitutor-Orbis/DISYS-MiniProject2/chittychat"
)

var username string

//client's lamport time
var lamportTime = chittychat.LamportTime{
	Time: 0,
}

func main() {
	// Creat a virtual RPC Client Connection on port  9080 WithInsecure (because  of http)
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}

	// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	client := chittychat.NewChittyChatClient(conn)

	fmt.Println("Type in your username")

	reader := bufio.NewReader(os.Stdin)
	clientMessage, err := reader.ReadString('\n')

	if err != nil {
		log.Fatalf("Failed to read from console")
	}

	username = strings.Trim(clientMessage, "\r\n")

	//Read user input in terminal
	go ReadFromTerminal(client)

	//read from server
	go PrintBroadcastsFromServer(client)

	for {
		t.Sleep(1000 * t.Hour)
	}
}

func ReadFromTerminal(client chittychat.ChittyChatClient) {
	for {
		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')

		if err != nil {
			log.Fatalf("Failed to read from console")
		}

		clientMessage = strings.Trim(clientMessage, "\r\n")

		publishRequest := chittychat.PublishRequest{
			User:    username,
			Message: clientMessage,
			Time:    lamportTime.Time,
		}

		PublishToServer(client, publishRequest)
	}
}

//call grpc method
func PublishToServer(client chittychat.ChittyChatClient, message chittychat.PublishRequest) {
	//increment lamport
	lamportTime.Time++
	message.Time = lamportTime.Time

	client.Publish(context.Background(), &message)
}

func PrintBroadcastsFromServer(client chittychat.ChittyChatClient) {

	//increment lamport
	lamportTime.Time++

	clientMessage := chittychat.SubscribeRequest{
		Username: username,
		Time:     lamportTime.Time,
	}

	//subscribe to the server
	//establish server-side streaming
	stream, err := client.Subscribe(context.Background(), &clientMessage)
	if err != nil {
		log.Fatalf("Error while opening stream %v", err)
	}

	for {
		messageToPrint, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		lamportTime.UpdateTime(messageToPrint.Time)
		lamportTime.Time++

		timeToString := strconv.FormatInt(int64(lamportTime.Time), 10)

		fmt.Println("["+timeToString+"]", messageToPrint.User+":", messageToPrint.Message)
	}
}
