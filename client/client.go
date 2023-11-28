package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	proto "grpc/GRPC"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ports = []string{":5001", ":5002", ":5003", ":5004"}
var id = flag.Int("id", 0, "ID of the client")

// ******   Methods for the client:
type Client struct {
	proto.AuctionServiceClient
	ID         int32
	Timestamp  int32
	Context    context.Context
	Servers    []proto.AuctionServiceClient
	ServerConn map[proto.AuctionServiceClient]*grpc.ClientConn
	CurrentIdx int // Index to keep track of the current primary server
}

func main() {
	//Set up logging
	f, err := os.OpenFile("../logs/clientlog.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v \n", err)
		fmt.Printf("error opening file: %v \n", err)
	}
	defer f.Close()

	log.SetOutput(f)

	flag.Parse()

	//Create a new client
	client := &Client{
		ID:        int32(*id),
		Timestamp: 0,
		Context:   context.Background(),
	}

	//Connect to the servers
	fmt.Print("Client joined Auction \n")
	client.ServerConn = make(map[proto.AuctionServiceClient]*grpc.ClientConn)
	client.joinAuction()
	defer client.closeAll()
	//Print server conn
	fmt.Printf("Client: %d, ServerConn: %v \n", client.ID, client.ServerConn)

	//Print the client struct
	fmt.Printf("ClientId: %d, Context: %v \n", client.ID, client.Context)
	//Start the auction
	client.runAuction()
}

func (client *Client) closeAll() {
	for _, c := range client.ServerConn {
		c.Close()
	}
}

func (client *Client) joinAuction() {
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client.ServerConn = make(map[proto.AuctionServiceClient]*grpc.ClientConn)
	for _, port := range ports {
		fmt.Printf("Client: %d trying to connect to server: %s \n", client.ID, port)
		log.Printf("Client: %d trying to connect to server: %s \n", client.ID, port)
		conn, err := grpc.DialContext(timeContext, fmt.Sprint(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("did not connect to %s: %s \n", port, err)
			continue
		}

		server := proto.NewAuctionServiceClient(conn)
		client.Servers = append(client.Servers, server)
		client.ServerConn[server] = conn
	}

	client.Context = context.Background()
	client.CurrentIdx = 0 // Start with the first server as primary
}

// *****    Methods for the auction functions:
func (client *Client) bid(amount int32) {
	newBid := &proto.BidInfo{
		Amount:    amount,
		BidderID:  client.ID,
		Timestamp: client.Timestamp,
	}
	var statusSuccess bool = false

	// Attempt to bid on all servers
	for idx, server := range client.Servers {
		ack, err := server.Bid(client.Context, newBid)
		if int(client.ID) == 69 {
			time.Sleep(time.Second * 5)
		}
		if err != nil {
			log.Printf("(Client %d) Error making bid on server %d: %s \n", client.ID, idx, err)
			continue // Move to the next server on error
		}

		if ack.Status == "fail" {
			fmt.Printf("Bid wasn't accepted \n")
			log.Printf("Client: %d recieved status from server %d: %s \n", client.ID, idx, ack.Status)
			log.Printf("Bid amount %d from client %d wasn't accepted \n", amount, client.ID)
			statusSuccess = false
			break // Stop bidding if the bid is not accepted
		} else if ack.Status == "success" {
			statusSuccess = true
		} else if ack.Status == "No new bids accepted. Auction is done" {
			fmt.Printf("Auction is done, bid is not registered \n")
			log.Printf("Client: %d recieved status from server %d: %s \n", client.ID, idx, ack.Status)
			statusSuccess = false
			break // Stop bidding if the auction is done
		}

		log.Printf("Client: %d recieved status from server %d: %s \n", client.ID, idx, ack.Status)
	}
	if statusSuccess {
		fmt.Printf("Bid was accepted \n")
		log.Printf("Bid amount %d from client %d was accepted \n", amount, client.ID)
	}
}

func (client *Client) result() {
	var currentResult *proto.CurrentResult
	var err error

	// Try to get the result from all available servers
	for _, server := range client.Servers {
		currentResult, err = server.Result(client.Context, &proto.ResultRequest{})
		if err != nil {
			log.Printf("(Client %d) Error getting result from server: %s\n", client.ID, err)
			continue // Move to the next server
		}
		// If successful, break the loop and print the result
		break
	}

	if err != nil {
		log.Fatal("All servers are unavailable")
		return
	}

	fmt.Printf("Auction status: %s - Highest bid: %d by participant: %d \n", currentResult.Status, currentResult.HighestBid, currentResult.HighestBidderID)
}

func (client *Client) runAuction() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter bid amount: \n")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "result" {
			client.result()
		} else {
			amount, err := strconv.Atoi(strings.TrimSpace(text))
			if err != nil {
				fmt.Println("Invalid input")
				continue
			}
			client.bid(int32(amount))
		}
	}
}
