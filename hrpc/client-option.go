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
	clientMaxConnectNum   = 10
	clientMaxMultiplexNum = 1000
	clientMaxActionNum    = clientMaxConnectNum * clientMaxMultiplexNum
)

type clientOptions struct {
	addr            string
	maxMsgSize      int
	timeout         time.Duration
	maxConnectNum   int
	maxMultiplexNum int
}

func newClientOptions() *clientOptions {
	return &clientOptions{
		maxMsgSize:      maxMsgSizeDef,
		timeout:         60 * time.Second,
		maxConnectNum:   clientMaxConnectNum,
		maxMultiplexNum: clientMaxMultiplexNum,
	}
}

type ClientOption interface {
	apply(*clientOptions)
}

type funcClientOption struct {
	f func(*clientOptions)
}

func (it *funcClientOption) apply(opts *clientOptions) {
	it.f(opts)
}

func ClientMaxMsgSize(v int) ClientOption {
	return &funcClientOption{
		f: func(opts *clientOptions) {
			if v < maxMsgSizeMin {
				v = maxMsgSizeMin
			} else if v > maxMsgSizeMax {
				v = maxMsgSizeMax
			}
			opts.maxMsgSize = v
		},
	}
}

func ClientTimeout(v time.Duration) ClientOption {
	return &funcClientOption{
		f: func(opts *clientOptions) {
			if v < 3e9 {
				v = 3e9
			} else if v > 120e9 {
				v = 120e9
			}
			opts.timeout = v
		},
	}
}

func ClientMaxConnectNum(v int) ClientOption {
	return &funcClientOption{
		f: func(opts *clientOptions) {
			if v < 1 {
				v = 1
			} else if v > clientMaxConnectNum {
				v = clientMaxConnectNum
			}
			opts.maxConnectNum = v
		},
	}
}

func ClientMaxMultiplexNum(v int) ClientOption {
	return &funcClientOption{
		f: func(opts *clientOptions) {
			if v < 10 {
				v = 10
			} else if v > clientMaxMultiplexNum {
				v = clientMaxMultiplexNum
			}
			opts.maxConnectNum = v
		},
	}
}
