package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"net"
	"os"
	"routes-api-gateway/pb"
	"routes-api-gateway/proxy"
	"strings"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse	, error) {
	log.Printf("Handling SayHello request [%v] with context %v", in, ctx)
	return &pb.PingResponse{Message: "You sent " + in.Message}, nil
}


var services map[string]string

func getTargetFromMethodName(methodName string) string{

	for k, v := range services{
		if strings.HasPrefix(methodName, k){
			return v
		}
	}

	return ""
}

func getServices()  {
	/*file, err := os.Open("/files/services.json")

	if err != nil{
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil{
		panic(err)
	}*/

	json.Unmarshal([]byte(os.Getenv("SERVICES")), &services)

}
func main() {
	fmt.Println("Starting Server")

	getServices()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil{
		log.Fatalf("Failed to listen %v", err)
	}
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		target := getTargetFromMethodName(fullMethodName)

		fmt.Printf("Routing %s to %s \n",fullMethodName, target)


		if target == ""{
			return ctx, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}

		conn, err := grpc.DialContext(ctx, target, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
		return ctx, conn, err
	}

	s := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	pb.RegisterPingServiceServer(s, &server{})

	if err := s.Serve(lis); err!=nil{
		log.Fatalf("Failed to serve: %v", err)
	}
}
