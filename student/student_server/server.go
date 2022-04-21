package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/SherzodAbdullajonov/grpc-golang/student/studentpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gopkg.in/mgo.v2/bson"
)

var collection *mongo.Collection

type server struct {
	studentpb.StudentServiceServer
}
type studentItem struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Department string             `bson:"department"`
	Course     int32              `bson:"course"`
	Address    string             `bson:"address"`
	Email      string             `bson:"email"`
}

func (*server) CreateStudent(ctx context.Context, req *studentpb.CreateStudentRequest) (*studentpb.CreateStudentResponse, error) {
	fmt.Println("Create student request")
	student := req.GetStudent()

	data := studentItem{
		FirstName:  student.GetFirstName(),
		LastName:   student.GetLastName(),
		Department: student.GetDepartment(),
		Address:    student.GetAddress(),
		Email:      student.GetEmail(),
		Course:     student.GetCourse(),
	}

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"),
		)
	}

	return &studentpb.CreateStudentResponse{
		Student: &studentpb.Student{
			Id:         oid.Hex(),
			FirstName:  student.GetFirstName(),
			LastName:   student.GetLastName(),
			Department: student.GetDepartment(),
			Address:    student.GetAddress(),
			Email:      student.GetEmail(),
			Course:     student.GetCourse(),
		},
	}, nil

}
func (*server) ReadStudent(ctx context.Context, req *studentpb.ReadStudentRequest) (*studentpb.ReadStudentResponse, error) {
	fmt.Println("Read student request")

	studentID := req.GetStudentId()
	oid, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	// create an empty struct
	data := &studentItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(ctx, filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	return &studentpb.ReadStudentResponse{
		Student: dataToStudentPb(data),
	}, nil
}

func dataToStudentPb(data *studentItem) *studentpb.Student {
	return &studentpb.Student{
		Id:         data.ID.Hex(),
		FirstName:  data.FirstName,
		LastName:   data.LastName,
		Address:    data.Address,
		Department: data.Department,
		Course:     data.Course,
		Email:      data.Email,
	}
}

func (*server) UpdateStudent(ctx context.Context, req *studentpb.UpdateStudentRequest) (*studentpb.UpdateStudentResponse, error) {
	fmt.Println("Update student request")
	student := req.GetStudent()
	oid, err := primitive.ObjectIDFromHex(student.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	// create an empty struct
	data := &studentItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(ctx, filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find student with specified ID: %v", err),
		)
	}

	// we update our internal struct

	data.FirstName = student.GetFirstName()
	data.LastName = student.GetLastName()
	data.Department = student.GetDepartment()
	data.Course = student.GetCourse()
	data.Address = student.GetAddress()
	data.Email = student.GetEmail()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in MongoDB: %v", updateErr),
		)
	}

	return &studentpb.UpdateStudentResponse{
		Student: dataToStudentPb(data),
	}, nil

}
func (*server) DeleteStudent(ctx context.Context, req *studentpb.DeleteStudentRequest) (*studentpb.DeleteStudentResponse, error) {
	fmt.Println("Delete student request")
	oid, err := primitive.ObjectIDFromHex(req.GetStudentId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	filter := bson.M{"_id": oid}

	res, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in MongoDB: %v", err),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find student in MongoDB: %v", err),
		)
	}

	return &studentpb.DeleteStudentResponse{StudentId: req.GetStudentId()}, nil
}

func (*server) ListStudent(_ *studentpb.ListStudentRequest, stream studentpb.StudentService_ListStudentServer) error {
	fmt.Println("List Student request")

	cur, err := collection.Find(context.Background(), primitive.D{{}})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cur.Close(context.Background()) // Should handle err
	for cur.Next(context.Background()) {
		data := &studentItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)

		}
		stream.Send(&studentpb.ListStudentResponse{Student: dataToStudentPb(data)}) // Should handle err
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Student Server started")
	collection = client.Database("mydb").Collection("student")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	studentpb.RegisterStudentServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	//Block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the server")
	lis.Close()
	fmt.Println("Closing MongoDb Connection")
	client.Disconnect(context.TODO())
	fmt.Println("End the program")
}
