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
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/req"
)

type clientContext struct {
	mu   sync.RWMutex
	sock mangos.Socket
	ctxQ chan mangos.Context
}

type clientActionContext struct {
	reqId  uint64
	req    []byte
	rep    []byte
	status chan uint8
	err    error
}

func newClientActionContext() *clientActionContext {
	return &clientActionContext{
		status: make(chan uint8, 2),
	}
}

type Client struct {
	mu      sync.RWMutex
	opts    *clientOptions
	conns   []*clientContext
	actions map[uint64]*clientActionContext
	sends   chan *clientActionContext
	idles   chan uint8
	close   bool
}

var (
	climu   sync.RWMutex
	clients = map[string]*Client{}
)

func NewClient(addr string, opt ...ClientOption) (*Client, error) {

	climu.Lock()
	defer climu.Unlock()

	opts := newClientOptions()

	opts.addr = addr

	c, ok := clients[opts.addr]
	if !ok {

		c = &Client{
			opts:    opts,
			conns:   []*clientContext{},
			actions: map[uint64]*clientActionContext{},
			sends:   make(chan *clientActionContext, clientMaxActionNum),
			idles:   make(chan uint8, clientMaxActionNum),
		}

		clients[opts.addr] = c

	} else {

		if c.opts.maxConnectNum < opts.maxConnectNum {
			c.opts.maxConnectNum = opts.maxConnectNum
		}
	}

	for len(c.conns) < c.opts.maxConnectNum {

		sock, err := c.newSocket()
		if err == nil {

			conn := &clientContext{
				sock: sock,
			}

			go c.worker(conn)

			c.conns = append(c.conns, conn)

		} else if err != nil && len(c.conns) == 0 {
			return nil, err
		}
	}

	return c, nil
}

func (it *Client) newSocket() (mangos.Socket, error) {

	sock, err := req.NewSocket()
	if err != nil {
		return nil, err
	}

	//
	sock.SetOption(mangos.OptionMaxRecvSize, it.opts.maxMsgSize+1024)

	//
	sock.SetOption(mangos.OptionSendDeadline, it.opts.timeout)
	sock.SetOption(mangos.OptionRecvDeadline, it.opts.timeout)

	if err = sock.Dial(msgTransport + "://" + it.opts.addr); err != nil {
		return nil, err
	}

	return sock, nil
}

func (it *Client) Invoke(ctx context.Context, methodName string, req, rep proto.Message) error {

	it.idles <- 1
	defer func() {
		<-it.idles
	}()

	var (
		meta   = newReqMeta(methodName)
		action = newClientActionContext()
	)

	action.reqId = meta.ReqId
	action.req = msgReqEncode(meta, req)

	it.mu.Lock()
	it.actions[meta.ReqId] = action
	it.mu.Unlock()

	it.sends <- action

	<-action.status

	if action.err != nil {
		return errors.New("client/invoke err " + action.err.Error())
	}

	_, err := msgRepDecode(action.rep, rep)

	return err
}

func (it *Client) worker(conn *clientContext) {

	conn.ctxQ = make(chan mangos.Context, clientMaxMultiplexNum)

	for len(conn.ctxQ) < clientMaxMultiplexNum {
		cctx, err := conn.sock.OpenContext()
		if err != nil {
			time.Sleep(1e9)
			continue
		}
		conn.ctxQ <- cctx
	}

	for !it.close {

		var (
			action = <-it.sends
			cctx   = <-conn.ctxQ
		)

		go it.callAction(conn, cctx, action)
	}
}

func (it *Client) callAction(conn *clientContext, cctx mangos.Context, action *clientActionContext) {

	defer func() {
		conn.ctxQ <- cctx
	}()

	var (
		reqMsg  = mangos.NewMessage(len(action.req))
		repMsg  *mangos.Message
		repMeta *RepMeta
		ok      bool
		err     error
	)

	// reqMsg.Body = action.req
	reqMsg.Body = append(reqMsg.Body, action.req...)

	if err = cctx.SendMsg(reqMsg); err == nil {

		if repMsg, err = cctx.RecvMsg(); err == nil {

			repMeta, err = msgRepDecode(repMsg.Body, nil)
		}
	}

	if err != nil {

		action.err = err
		action.status <- 1

	} else if repMeta != nil {

		it.mu.Lock()
		action, ok = it.actions[repMeta.ReqId]
		if ok {
			delete(it.actions, repMeta.ReqId)
		}
		it.mu.Unlock()

		if action != nil {
			action.rep = repMsg.Body
			action.status <- 1
		}
	}
}

func (it *Client) Close() error {

	climu.Lock()
	defer climu.Unlock()

	if it.close {
		return nil
	}
	it.close = true

	for i := 0; i < 60; i++ {

		if len(it.actions) == 0 && len(it.sends) == 0 {
			break
		}

		time.Sleep(1e3)
	}

	for _, c := range it.conns {
		c.sock.Close()
	}

	return nil
}
