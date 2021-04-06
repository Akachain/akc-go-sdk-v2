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

package mock

import (
	"errors"
	. "github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/util/couchdb"
)

// AkcQueryIterator inherits StateQueryIterator to simulate how the peer handle query string response
type AkcQueryIterator struct {
	data       []*couchdb.QueryResult
	currentLoc int
	*StateQueryIterator
}

func (it *AkcQueryIterator) HasNext() bool {
	return it.currentLoc < len(it.data)
}

func (it *AkcQueryIterator) Length() int {
	return len(it.data)
}

func (it *AkcQueryIterator) Next() (*queryresult.KV, error) {
	var kv = new(queryresult.KV)

	if !it.HasNext() {
		return nil, errors.New("there is no other item in the iterator")
	}

	item := it.data[it.currentLoc]
	it.currentLoc++

	if item == nil {
		return nil, errors.New("empty query result")
	}

	kv.Value = item.Value

	return kv, nil
}

func (it *AkcQueryIterator) Close() error {
	return nil
}

// FromResultsIterator provides a way of converting ResultsIterator into StateQueryIterator
func FromResultsIterator(rit statedb.ResultsIterator) (*AkcQueryIterator, error) {
	// Init the result iterator
	rawData := make([]*couchdb.QueryResult, 0)
	iterator := &AkcQueryIterator{data: rawData, currentLoc: 0}

	// Fill it with raw data
	for {
		member, er := rit.Next()

		if er != nil {
			return nil, er
		}

		// no more member
		if member == nil {
			break
		}

		// convert VersionedKV to QueryResult
		z := member.(*statedb.VersionedKV)
		r := new(couchdb.QueryResult)
		r.ID = z.Key
		r.Value = z.Value
		rawData = append(rawData, r)
	}

	iterator.data = rawData
	iterator.currentLoc = 0
	return iterator, nil
}
