package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/SherzodAbdullajonov/grpc-golang/greet/greetpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Print("Hello I'm client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	//fmt.Printf("Created client: %f", c)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBidiStreaming(c)

}
func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println(" Starting to Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sherzod",
			LastName:  "Abdullajonov",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}
	log.Printf("Response form Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println(" Starting to do a Server Streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sherzod",
			LastName:  "Abdullajonov",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break

		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Responce form GreetManyTimes: %v", msg.GetResult())
	}
}
func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Startingt to do a Client Streaming RPC ...")
	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sherzod",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Shokhrukh",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Khurshid",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jakhongir",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Shakhzod",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jaloldin",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Mirkamol",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Muhammadyusuf",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while recieving response form LongGreet: %v", err)
	}
	fmt.Printf("LongGreet Response: %v\n", res)
}
func doBidiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Startingt to do a Client Streaming RPC ...")
	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Sherzod",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Shokhrukh",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Khurshid",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jakhongir",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Shakhzod",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jaloldin",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Mirkamol",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Muhammadyusuf",
			},
		},
	}
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
	}
	waitc := make(chan struct{})
	// go routine function to send a branch of massages to the client
	go func() {
		for _, req := range requests {
			fmt.Printf("Sending messages: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// go routine funciton to recieve a branch of  meesages
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while recieving: %v", err)
				break
			}
			fmt.Printf("Recieved: %v\n", res.GetResult())
		}
		close(waitc)
	}()
	<-waitc
}
