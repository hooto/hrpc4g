// Copyright 2020 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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
	"errors"
	"fmt"
	"time"

	"github.com/hooto/hrpc4g/hrpc"

	rpcbench "github.com/hooto/rpcbench/v1"
)

func main() {

	b := rpcbench.NewBench(new(Bench), new(BenchServiceServer))

	if err := b.Run(); err != nil {
		panic(err)
	}
}

type BenchServiceServer struct {
	server  *hrpc.Server
	connNum int
}

func (s BenchServiceServer) UnaryCall(
	ctx context.Context,
	req *rpcbench.BenchRequest,
) (*rpcbench.BenchReply, error) {

	if req.ReplySpec == nil {
		return nil, errors.New("ReplySpec not found")
	}

	if req.ReplySpec.WorkTime > 0 {
		time.Sleep(time.Microsecond * time.Duration(req.ReplySpec.WorkTime))
	}

	return &rpcbench.BenchReply{
		Id: req.Id,
		Payload: &rpcbench.BenchPayload{
			Body: rpcbench.RandValue(int(req.ReplySpec.Size)),
		},
	}, nil
}

type Bench struct {
	server        *hrpc.Server
	client        *hrpc.Client
	serviceServer *BenchServiceServer
}

type BenchClientConn struct {
	client *hrpc.Client
}

func (it *Bench) ServerStart(cfg *rpcbench.Config) error {

	var err error

	it.server, err = hrpc.NewServer(fmt.Sprintf("127.0.0.1:%d", cfg.ServerPort))
	if err != nil {
		return err
	}

	ss := &BenchServiceServer{
		server: it.server,
	}

	it.server.RegisterService(
		rpcbench.HrpcBenchmarkServiceDesc, ss)

	go it.server.Serve()

	return nil
}

func (it *Bench) ClientConn(opts *rpcbench.ClientConnOptions) (rpcbench.ClientConnInterface, error) {

	client, err := hrpc.NewClient(opts.Addr,
		hrpc.ClientMaxConnectNum(opts.ConnNum))
	if err != nil {
		return nil, err
	}

	return &BenchClientConn{
		client: client,
	}, nil
}

func (it *BenchClientConn) UnaryCall(
	ctx context.Context,
	req *rpcbench.BenchRequest) (*rpcbench.BenchReply, error) {

	return rpcbench.HrpcBenchmarkClient(it.client).UnaryCall(ctx, req)
}
