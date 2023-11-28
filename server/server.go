package main

import (
	"context"
	"flag"
	"fmt"
	proto "grpc/GRPC"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

var server *Server
var ports = []string{":5001", ":5002", ":5003", ":5004"}
var serverName = flag.String("name", "serverX", "Name of the server")
var min = flag.Int("min", 5, "amount of minutes the auction is running")

type Server struct {
	proto.UnimplementedAuctionServiceServer
	name string
	port string

	Timestamp      int32
	HighestBid     int32
	BidderMap      map[int32]int32
	AuctionOngoing bool
}

func main() {
	//Set up logging
	f, err := os.OpenFile("../logs/serverlog.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v \n", err)
		fmt.Printf("error opening file: %v \n", err)
	}
	defer f.Close()

	log.SetOutput(f)

	flag.Parse()
	startServer(ports[:])
	fmt.Print("This place is never reached \n")

	//loop to make sure the server doesn't close:
	for {
		time.Sleep(5000) //sleep for 5 seconds
	}
}

func startServer(ports []string) {

	//Create listener
	listener, err := net.Listen("tcp", ports[0])
	if err != nil {
		if len(ports) > 1 {
			startServer(ports[1:])
		} else {
			log.Fatalf("error creating the server %v \n", err)
			fmt.Printf("error creating the server %v \n", err)
		}
	}

	//Create the server and register as a GRPC server
	server = newServer(ports[0])
	grpcServer := grpc.NewServer()
	proto.RegisterAuctionServiceServer(grpcServer, server)

	//Handle SIGINT and SIGTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Printf("%s Received termination signal. Shutting down...\n", server.name)
		log.Printf("%s Received termination signal. Shutting down...\n", server.name)
		os.Exit(0)
	}()

	//Run auction
	go runAuction()

	//Start the server
	fmt.Printf("Starting server on port %s \n", ports[0])
	log.Printf("Starting server on port %s \n", ports[0])
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("error creating the server %v \n", err)
		fmt.Printf("error creating the server %v \n", err)
	}
	log.Println("Server closed.")
	fmt.Println("Server closed.")

}

func newServer(port string) *Server {
	return &Server{
		name:           fmt.Sprintf("server%s", port),
		port:           port,
		Timestamp:      0,
		AuctionOngoing: false,
		HighestBid:     0,
		BidderMap:      make(map[int32]int32),
	}
}

func runAuction() {
	server.AuctionOngoing = true
	time.Sleep(time.Minute * time.Duration(*min)) //sleep for 60 seconds (or flag min)
	server.AuctionOngoing = false
	log.Printf("Auction has ended \n The winner is: client %d who bid %d \n", server.BidderMap[server.HighestBid], server.HighestBid)
	fmt.Printf("Auction has ended \n The winner is: client %d who bid %d \n", server.BidderMap[server.HighestBid], server.HighestBid)
}

func (s *Server) Bid(contxt context.Context, bidInfo *proto.BidInfo) (*proto.Ack, error) {
	if s.AuctionOngoing {
		if bidInfo.Amount > s.HighestBid {
			s.HighestBid = bidInfo.Amount
			s.BidderMap[bidInfo.Amount] = bidInfo.BidderID
			log.Printf("(%s) Bidder %d has succesfully bid %d \n", s.name, bidInfo.BidderID, bidInfo.Amount)
			fmt.Printf("(%s) Bidder %d has succesfully bid %d \n", s.name, bidInfo.BidderID, bidInfo.Amount)
			return &proto.Ack{Status: "success"}, nil
		} else {
			log.Printf("(%s) Bidder %d attempted to bid %d - less than highest bid: %d \n", s.name, bidInfo.BidderID, bidInfo.Amount, s.HighestBid)
			fmt.Printf("(%s) Bidder %d attempted to bid %d - less than highest bid: %d \n", s.name, bidInfo.BidderID, bidInfo.Amount, s.HighestBid)
			return &proto.Ack{Status: "fail"}, nil
		}
	} else {
		return &proto.Ack{Status: "No new bids accepted. Auction is done"}, nil
	}

}

func (s *Server) Result(context.Context, *proto.ResultRequest) (currentResult *proto.CurrentResult, err error) {

	var status string
	if s.AuctionOngoing {
		status = "in progress"
	} else {
		status = "finished"
	}

	return &proto.CurrentResult{
		Status:          status,
		HighestBid:      s.HighestBid,
		HighestBidderID: s.BidderMap[s.HighestBid],
	}, err
}
