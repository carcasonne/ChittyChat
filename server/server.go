package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"

	chittychat "github.com/Restitutor-Orbis/DISYS-MiniProject2/chittychat"

	"google.golang.org/grpc"
)

type Server struct {
	chittychat.UnimplementedChittyChatServer
}

//slice of all known connections
var sliceOfStreams []chittychat.ChittyChat_SubscribeServer

//server's lamport time
//init to 0
var lamportTime = chittychat.LamportTime{
	Time: 0,
}

func main() {

	//log to file, taken from https://dev.to/gholami1313/saving-log-messages-to-a-custom-log-file-in-golang-ce5
	LOG_FILE := "log.txt"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(GetTimeAsString(), err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("-----------------------------------")

	log.Println(GetTimeAsString(), "Logging to custom file")

	// Create listener tcp on port 9080
	list, err := net.Listen("tcp", ":9081")
	if err != nil {
		log.Fatalf(GetTimeAsString(), "Failed to listen on port 9081: %v", err)
	}

	grpcServer := grpc.NewServer()
	chittychat.RegisterChittyChatServer(grpcServer, &Server{})

	log.Println(GetTimeAsString(), "Server is set up on port 9081")

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf(GetTimeAsString(), "failed to server %v", err)
	}

	//grpc listen and serve
	err = grpcServer.Serve(list)
	if err != nil {
		log.Fatalf(GetTimeAsString(), "Failed to start gRPC Server :: %v", err)
	}
}

func (s *Server) Publish(ctx context.Context, in *chittychat.PublishRequest) (*chittychat.PublishReply, error) {
	//increment lamport time
	lamportTime.UpdateTime(in.Time)
	lamportTime.Mutex.Lock()
	lamportTime.Time++
	lamportTime.Mutex.Unlock()

	log.Println(GetTimeAsString(), "Received PublishRequest from", in.User)

	broadcastReply := chittychat.SubscribeReply{
		User:    in.User,
		Message: in.Message,
		Time:    lamportTime.Time,
	}

	BroadcastToAllClients(&broadcastReply)

	return &chittychat.PublishReply{}, nil
}

func (s *Server) Subscribe(in *chittychat.SubscribeRequest, subscriptionServer chittychat.ChittyChat_SubscribeServer) error {

	//increment lamport time
	//update against client lamport time
	lamportTime.UpdateTime(in.Time)
	lamportTime.Mutex.Lock()
	lamportTime.Time++
	lamportTime.Mutex.Unlock()

	message := chittychat.SubscribeReply{
		User:    in.Username,
		Message: "has joined the chat",
		Time:    lamportTime.Time,
	}

	//save this stream
	//this should maybe be handled in a separate go routine, to prevent the server from being killed off?
	sliceOfStreams = append(sliceOfStreams, subscriptionServer)

	log.Println(GetTimeAsString(), "Added new client stream to server")

	BroadcastToAllClients(&message)

	//prevent function from terminating
	//keeps the stream connection alive
	for {
		select {
		case <-subscriptionServer.Context().Done():

			//increment time as "recieved" message from client that it disconnected
			lamportTime.Mutex.Lock()
			lamportTime.Time++
			lamportTime.Mutex.Unlock()

			broadcastReply := chittychat.SubscribeReply{
				User:    message.User,
				Message: "has left the chat",
				Time:    lamportTime.Time,
			}

			log.Println(GetTimeAsString(), message.User+" has disconnected")

			BroadcastToAllClients(&broadcastReply)

			/* for _, element := range sliceOfStreams {
				if element == subscriptionServer {
					element = nil
				}
			} */

			return nil
		}
	}
}

func BroadcastToAllClients(message *chittychat.SubscribeReply) {

	//increment lamport time
	lamportTime.Mutex.Lock()
	lamportTime.Time++
	lamportTime.Mutex.Unlock()
	message.Time = lamportTime.Time
	log.Println(GetTimeAsString(), "Broadcasting to all clients")

	//send message to every known stream
	for _, element := range sliceOfStreams {
		if element != nil {
			element.Send(message)
		}
	}
}

//doesn't require lamport increment
func GetTimeAsString() string {
	timeToString := strconv.FormatInt(int64(lamportTime.Time), 10)
	return "[" + timeToString + "]"
}
