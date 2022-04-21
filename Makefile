create:
	protoc greet/greetpb/greet.proto --go_out=plugins=grpc:. 
	protoc calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.
	protoc student/studentpb/student.proto --go_out=plugins=grpc:.