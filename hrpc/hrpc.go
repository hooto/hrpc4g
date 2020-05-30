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

//go:generate protoc --go_out=plugins=grpc:. hrpc.proto

import (
	"context"

	"github.com/golang/protobuf/proto"
)

const (
	ErrCodeBadArgument  = 2
	ErrCodeClientError  = 3
	ErrCodeServerError  = 4
	ErrCodeAccessDenied = 5
	ErrCodeDataNotFound = 6
	ErrCodeDataConflict = 7
)

type serviceMethodHandler func(srv interface{},
	ctx context.Context, inb []byte) (proto.Message, error)

type ServiceMethodDesc struct {
	MethodName string
	Handler    serviceMethodHandler
}

type ServiceDesc struct {
	ServiceName string
	HandlerType interface{}
	Methods     []ServiceMethodDesc
}

type serviceMethodItem struct {
	service     interface{}
	ServiceName string
	MethodName  string
	Handler     serviceMethodHandler
}
