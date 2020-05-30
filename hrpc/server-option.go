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
	"time"
)

const (
	msgTransport  = "tcp"
	maxMsgSizeDef = 8 * 1024 * 1024
	maxMsgSizeMin = 1 * 1024 * 1024
	maxMsgSizeMax = 64 * 1024 * 1024
	msgTimeout    = 60 * time.Second
)

type serverOptions struct {
	Addr       string
	MaxMsgSize int
	Timeout    time.Duration
}

func newServerOptions() *serverOptions {
	return &serverOptions{
		MaxMsgSize: maxMsgSizeDef,
	}
}

type ServerOption interface {
	apply(*serverOptions)
}

type funcServerOption struct {
	f func(*serverOptions)
}

func (it *funcServerOption) apply(opts *serverOptions) {
	it.f(opts)
}

func MaxMsgSize(v int) ServerOption {
	return &funcServerOption{
		f: func(opts *serverOptions) {
			if v < maxMsgSizeMin {
				v = maxMsgSizeMin
			} else if v > maxMsgSizeMax {
				v = maxMsgSizeMax
			}
			opts.MaxMsgSize = v
		},
	}
}

func ConnectionTimeout(v time.Duration) ServerOption {
	return &funcServerOption{
		f: func(opts *serverOptions) {
			if v < 1e9 {
				v = 1e9
			}
			opts.Timeout = v
		},
	}
}
