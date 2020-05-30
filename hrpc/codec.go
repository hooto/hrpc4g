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
	"errors"
	"math/rand"

	"github.com/golang/protobuf/proto"
)

const (
	MsgCodecVer200 uint8 = 0x01
)

func newReqMeta(fullMethodName string) *ReqMeta {
	return &ReqMeta{
		ReqId:      rand.Uint64(),
		FullMethod: fullMethodName,
	}
}

func msgReqEncode(metaEntry *ReqMeta, msgEntry proto.Message) []byte {

	buf := []byte{MsgCodecVer200, 0x00}

	metaEnc, _ := proto.Marshal(metaEntry)
	if n := len(metaEnc); n > 0 {
		buf[1] = uint8(n)
		buf = append(buf, metaEnc...)
	}

	msgEnc, _ := proto.Marshal(msgEntry)
	if len(msgEnc) > 0 {
		buf = append(buf, msgEnc...)
	}

	return buf
}

func msgReqDecode(buf []byte) (*ReqMeta, []byte, error) {

	if len(buf) < 2 {
		return nil, nil, errors.New("invalid codec")
	}

	switch buf[0] {

	case MsgCodecVer200:
		mSize := int(buf[1])
		if (2 + mSize) >= len(buf) {
			return nil, nil, errors.New("invalid codec")
		}
		var meta ReqMeta
		if err := proto.Unmarshal(buf[2:2+mSize], &meta); err != nil {
			return nil, nil, err
		}
		return &meta, buf[mSize+2:], nil
	}

	return nil, nil, errors.New("invalid codec")
}

func msgRepEncode(meta *RepMeta, err error, msgEntry proto.Message) []byte {

	if meta == nil {
		meta = &RepMeta{}
	}

	if err != nil {
		meta.ErrorCode = ErrCodeServerError
		meta.ErrorMessage = err.Error()
	}

	var msgEnc []byte
	if msgEntry != nil {
		msgEnc, _ = proto.Marshal(msgEntry)
		if len(msgEnc) > 0 {
			meta.BodySize = int32(len(msgEnc))
		}
	}

	buf := []byte{MsgCodecVer200, 0x00}

	metaEnc, _ := proto.Marshal(meta)
	if n := len(metaEnc); n > 0 {
		buf[1] = uint8(n)
		buf = append(buf, metaEnc...)
	}

	if len(msgEnc) > 0 {
		buf = append(buf, msgEnc...)
	}

	return buf
}

func msgRepDecode(buf []byte, rep proto.Message) (*RepMeta, error) {

	if len(buf) < 2 {
		return nil, errors.New("invalid codec (raw size)")
	}

	switch buf[0] {

	case MsgCodecVer200:

		mSize := int(buf[1])
		if (2 + mSize) > len(buf) {
			return nil, errors.New("invalid codec (meta size)")
		}

		var meta RepMeta
		if err := proto.Unmarshal(buf[2:2+mSize], &meta); err != nil {
			return nil, err
		}

		if meta.BodySize > 0 {
			if (2 + mSize + int(meta.BodySize)) != len(buf) {
				return nil, errors.New("invalid codec (msg size)")
			}
			if rep != nil {
				if err := proto.Unmarshal(buf[2+mSize:], rep); err != nil {
					return nil, err
				}
			}
		}

		return &meta, nil
	}

	return nil, errors.New("invalid codec (version)")
}
