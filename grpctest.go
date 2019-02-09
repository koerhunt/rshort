package main

import (
"context"
"log"
	"os"
	"time"

"google.golang.org/grpc"
pb "github.com/koerhunt/rshort/grpc"
)

const (
	address     = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRshorterClient(conn)

	// Contact the server and print out its response.
	key:= "g1"
	url := "https://google.com/1"

	if len(os.Args) > 2 {
		key = os.Args[1]
		url = os.Args[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := c.CutURL(ctx, &pb.CutUrlRequest{Key: key,Url: url})
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	log.Printf("Status: %s", r.Status)
	log.Printf("Data: %s", r.Data)

}

