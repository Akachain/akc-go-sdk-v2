// Copyright (c) 2021 akachain
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package common

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const (
	OK    = 200
	ERROR = 500
)

const (
	SUCCESS = "200"
	ERR1    = "AKC0001"
	ERR2    = "AKC0002"
	ERR3    = "AKC0003"
	ERR4    = "AKC0004"
	ERR5    = "AKC0005"
	ERR6    = "AKC0006"
	ERR7    = "AKC0007"
	ERR8    = "AKC0008"
	ERR9    = "AKC0009"
	ERR10   = "AKC0010"
	ERR11   = "AKC0011"
	ERR12   = "AKC0012"
	ERR13   = "AKC0013"
	ERR14   = "AKC0014"
	ERR15   = "AKC0015"
	ERR16   = "AKC0016"
	ERR17   = "AKC0017"
	ERR18   = "AKC0018"
	ERR19   = "AKC0019"
	ERR20   = "AKC0020"
)

var ResCodeDict = map[string]string{
	"200":     "OK",
	"AKC0001": "Cannot Update User Information!",
	"AKC0002": "Incorrect number of arguments!",
	"AKC0003": "Convert Json fail!",
	"AKC0004": "Get data fail!",
	"AKC0005": "Insert data fail!",
	"AKC0006": "Public key error!",
	"AKC0007": "ParsePKCS1PublicKey error!",
	"AKC0008": "Verify error!",
	"AKC0009": "Only signed once!",
	"AKC0010": "Not Enough Quorum",
	"AKC0011": "Only Commit once!",
	"AKC0012": "Proposal Commit not exist!",
	"AKC0013": "Proposal ID not exist!",
	"AKC0014": "Admin ID not exist!",
	"AKC0015": "Admin ID not Active!",
	"AKC0016": "Proposal Rejected!",
	"AKC0017": "You have confirmed you cannot reject!",
	"AKC0018": "Only reject once!",
	"AKC0019": "This ApprovalID can't be empty",
	"AKC0020": "Approver not Active!",
}

type InvokeResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Rows   string `json:"rows"`
}

type ResponseSuccess struct {
	ResCode string
	Msg     string
	Payload string
}

type ResponseError struct {
	ResCode string
	Msg     string
}

func RespondSuccess(res ResponseSuccess) pb.Response {
	if res.Payload == "" {
		res.Payload = "[]"
	}
	return pb.Response{
		Status:  OK,
		Payload: []byte(res.Payload),
		Message: res.Payload,
	}
}

func RespondError(err ResponseError) pb.Response {
	msg := "{\"status\":\"" + err.ResCode + "\", \"msg\":\"" + err.Msg + "\"}"
	return pb.Response{
		Status:  ERROR,
		Message: msg,
	}
}
