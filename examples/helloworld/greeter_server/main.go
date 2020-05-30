// Copyright 2020 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log"

	pb "github.com/hooto/hrpc4g/examples/helloworld/helloworld"
	"github.com/hooto/hrpc4g/hrpc"
)

var (
	argAddr = "127.0.0.1:13301"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", req.Name)
	return &pb.HelloReply{Message: req.Name}, nil
}

func main() {

	s, err := hrpc.NewServer(argAddr)
	if err != nil {
		log.Fatalf("failed to create server : %v", err)
	}

	s.RegisterService(pb.HrpcGreeterServiceDesc, new(server))

	if err := s.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
