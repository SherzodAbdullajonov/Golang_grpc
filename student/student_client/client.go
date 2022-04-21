package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/SherzodAbdullajonov/grpc-golang/student/studentpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Student Client")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close() // Maybe this should be in a separate function and the error handled?

	c := studentpb.NewStudentServiceClient(cc)

	// create Student
	fmt.Println("Creating the student")
	student := &studentpb.Student{
		FirstName:  "Sherzod",
		LastName:   "Abdullajonov",
		Address:    "Fergana",
		Department: "Socie",
		Course:     4,
		Email:      "sabdullajonov@gmail.com",
	}
	createStudentRes, err := c.CreateStudent(context.Background(), &studentpb.CreateStudentRequest{Student: student})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Blog has been created: %v", createStudentRes)
	studentID := createStudentRes.GetStudent().GetId()

	// read Student
	fmt.Println("Reading the blog")

	_, err2 := c.ReadStudent(context.Background(), &studentpb.ReadStudentRequest{StudentId: "5bdc29e661b75adcac496cf4"})
	if err2 != nil {
		fmt.Printf("Error happened while reading: %v \n", err2)
	}

	readStudentReq := &studentpb.ReadStudentRequest{StudentId: studentID}
	readStudentRes, readStudentErr := c.ReadStudent(context.Background(), readStudentReq)
	if readStudentErr != nil {
		fmt.Printf("Error happened while reading: %v \n", readStudentErr)
	}

	fmt.Printf("Student was read: %v \n", readStudentRes)
	// update Student
	newStudent := &studentpb.Student{
		Id:         studentID,
		FirstName:  "Khurshid",
		LastName:   "Kabilov",
		Address:    "Fergana",
		Email:      "kabilov@gmail.com",
		Course:     4,
		Department: "Socie",
	}
	updateRes, updateErr := c.UpdateStudent(context.Background(), &studentpb.UpdateStudentRequest{Student: newStudent})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Student was updated: %v\n", updateRes)
	// delete Student
	deleteRes, deleteErr := c.DeleteStudent(context.Background(), &studentpb.DeleteStudentRequest{StudentId: studentID})

	if deleteErr != nil {
		fmt.Printf("Error happened while deleting: %v \n", deleteErr)
	}
	fmt.Printf("Student was deleted: %v \n", deleteRes)

	// list Blogs

	stream, err := c.ListStudent(context.Background(), &studentpb.ListStudentRequest{})
	if err != nil {
		log.Fatalf("error while calling ListBlog RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetStudent())
	}
}
