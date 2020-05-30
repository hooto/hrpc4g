// Copyright 2020 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package hrpc

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/rep"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

var (
	serverWorkerNumMax = 100
	serverActionNumMax = 100000
)

type Server struct {
	mu       sync.RWMutex
	opts     *serverOptions
	sock     mangos.Socket
	services map[string]*serviceMethodItem
	idles    chan int
	running  bool
	close    bool
}

type serverContext struct {
	ctx mangos.Context
}

type serverActionContext struct {
	ctx mangos.Context
	req []byte
}

func NewServer(addr string, opt ...ServerOption) (*Server, error) {

	sock, err := rep.NewSocket()
	if err != nil {
		return nil, err
	}

	addr = msgTransport + "://" + addr
	if err = sock.Listen(addr); err != nil {
		return nil, err
	}

	opts := newServerOptions()
	for _, v := range opt {
		v.apply(opts)
	}

	//
	sock.SetOption(mangos.OptionMaxRecvSize, opts.MaxMsgSize+1024)

	//
	sock.SetOption(mangos.OptionSendDeadline, opts.Timeout)
	sock.SetOption(mangos.OptionRecvDeadline, opts.Timeout)

	return &Server{
		opts:     opts,
		sock:     sock,
		services: map[string]*serviceMethodItem{},
		idles:    make(chan int, serverActionNumMax),
		close:    false,
	}, nil
}

func (it *Server) RegisterService(sd *ServiceDesc, ss interface{}) error {

	it.mu.Lock()
	defer it.mu.Unlock()

	var (
		ht = reflect.TypeOf(sd.HandlerType).Elem()
		st = reflect.TypeOf(ss)
	)

	if !st.Implements(ht) {
		return errors.New("invalid hrpc service")
	}

	for _, v := range sd.Methods {

		it.services[sd.ServiceName+"/"+v.MethodName] = &serviceMethodItem{
			service:     ss,
			ServiceName: sd.ServiceName,
			MethodName:  v.MethodName,
			Handler:     v.Handler,
		}
	}

	return nil
}

func (it *Server) Serve() error {

	it.mu.Lock()
	if it.running {
		it.mu.Unlock()
		return nil
	}
	it.running = true
	it.mu.Unlock()

	for !it.close {

		sctx, err := it.sock.OpenContext()
		if err != nil {
			time.Sleep(1e3)
			continue
		}

		sctx.SetOption(mangos.OptionMaxRecvSize, it.opts.MaxMsgSize+1024)

		if msg, err := sctx.Recv(); err == nil {

			action := &serverActionContext{
				ctx: sctx,
				req: msg,
			}

			it.idles <- 1
			go func() {
				it.runAction(action)
				_ = <-it.idles
				sctx.Close()
			}()

		} else {
			sctx.Close()
		}
	}

	return nil
}

func (it *Server) runAction(action *serverActionContext) {

	var (
		rep                   proto.Message
		reqMeta, reqData, err = msgReqDecode(action.req)
	)

	if err == nil {

		it.mu.RLock()
		m, ok := it.services[reqMeta.FullMethod]
		it.mu.RUnlock()

		if !ok {
			err = errors.New("no method found")
		} else {
			var ctx context.Context
			rep, err = m.Handler(m.service, ctx, reqData)
		}
	}

	repData := msgRepEncode(&RepMeta{
		ReqId: reqMeta.ReqId,
	}, err, rep)

	repMsg := mangos.NewMessage(len(repData))
	repMsg.Body = repData

	action.ctx.SendMsg(repMsg)
}

func (it *Server) Close() error {
	it.close = true
	time.Sleep(1e9)
	return it.sock.Close()
}
