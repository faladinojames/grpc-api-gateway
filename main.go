package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"net"
	"os"
	"routes-api-gateway/pb"
	"routes-api-gateway/proxy"
	"strings"
)
import _ "github.com/joho/godotenv/autoload"

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("Handling SayHello request [%v] with context %v", in, ctx)
	return &pb.PingResponse{Message: "You sent " + in.Message}, nil
}

var services map[string]interface{}

func getTargetFromMethodName(methodName string) (string, bool) {

	for k, v := range services {
		if strings.HasPrefix(methodName, k) {
			config := v.(map[string]interface{})
			return config["address"].(string), config["requiresAuth"].(bool)
		}
	}

	return "", false
}

func getServices() {
	file, err := os.Open("/files/services.json")

	if err != nil {
		//panic(err)
		file, err = os.Open("./local-services.json")

		if err != nil {
			panic(err)
		}
	}

	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(bytes, &services)

}
func main() {
	fmt.Println("Starting Server")

	getServices()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		target, requiresAuth := getTargetFromMethodName(fullMethodName)

		fmt.Printf("Routing %s to %s \n", fullMethodName, target)

		if target == "" {
			return ctx, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		}

		if requiresAuth {
			md, _ := metadata.FromIncomingContext(ctx)
			sessionId := ""
			if len(md.Get("x-session-id")) > 0 {
				sessionId = md.Get("x-session-id")[0]
			}

			if sessionId == "" {
				return ctx, nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
			} else {
				userData, err := getUserFromSession(sessionId)

				if err != nil {
					return ctx, nil, grpc.Errorf(codes.Unauthenticated, "Invalid Authentication")
				}
				forwardMD := metadata.Pairs("x-user", userData)
				ctx = metadata.NewIncomingContext(ctx, metadata.Join(md, forwardMD))
			}
		}

		conn, err := grpc.DialContext(ctx, target, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
		return ctx, conn, err
	}

	s := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

	pb.RegisterPingServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
