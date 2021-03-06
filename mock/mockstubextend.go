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
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/spf13/viper"
	"strings"
	"unicode/utf8"

	logging "github.com/op/go-logging"
)

const (
	maxUnicodeRuneValue = utf8.MaxRune //U+10FFFF - maximum (and unallocated) code point
)

// Logger for the shim package.
var mockLogger = logging.MustGetLogger("mockStubExtend")

// MockStubExtend provides composition class for MockStub as some of the mockstub methods are not implemented
type MockStubExtend struct {
	args      [][]byte        // this is private in MockStub
	cc        shim.Chaincode  // this is private in MockStub
	CouchDB   bool            // if we use couchDB
	DbHandler *CouchDBHandler // if we use couchDB
	*shimtest.MockStub
}

// GetQueryResult overrides the same function in MockStub
// that did not implement anything.
func (stub *MockStubExtend) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	// Query data from couchDB
	raw, err := stub.DbHandler.QueryDocument(query)
	if err != nil {
		return nil, err
	}
	return FromResultsIterator(raw)
}

// GetQueryResultWithPagination overrides the same function in MockStub
// that did not implement anything.
func (stub *MockStubExtend) GetQueryResultWithPagination(query string, pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {

	raw, er := stub.DbHandler.QueryDocumentWithPagination(query, pageSize, bookmark)
	if er != nil {
		return nil, nil, er
	}

	iterator, er := FromResultsIterator(raw)
	if er != nil {
		return nil, nil, er
	}

	bm := raw.(statedb.QueryResultsIterator).GetBookmarkAndClose()
	queryResponse := &pb.QueryResponseMetadata{FetchedRecordsCount: int32(iterator.Length()), Bookmark: bm}

	return iterator, queryResponse, nil
}

// NewMockStubExtend creates a new MockStubExtend object.
// You must set config path to the directory contains core.yaml file.
func NewMockStubExtend(stub *shimtest.MockStub, cc shim.Chaincode, configPath string) *MockStubExtend {
	s := new(MockStubExtend)
	s.MockStub = stub
	s.cc = cc
	s.CouchDB = false
	viper.SetConfigName("core")
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
	return s
}

// SetCouchDBConfiguration sets the couchdb configuration with appropriate database handler
func (stub *MockStubExtend) SetCouchDBConfiguration(handler *CouchDBHandler) {
	stub.CouchDB = true
	stub.DbHandler = handler
}

// MockInvoke Override this function from MockStub
func (stub *MockStubExtend) MockInvoke(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

// MockInit Override this function from MockStub
func (stub *MockStubExtend) MockInit(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

// GetFunctionAndParameters Override this function from MockStub
func (stub *MockStubExtend) GetFunctionAndParameters() (function string, params []string) {
	allargs := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}

// GetStringArgs override this function from MockStub
func (stub *MockStubExtend) GetStringArgs() []string {
	strargs := make([]string, 0, len(stub.args))
	for _, barg := range stub.args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

// PutState writes the specified `value` and `key` into the ledger.
func (stub *MockStubExtend) PutState(key string, value []byte) error {
	// In case we are using CouchDB, we store the value document in the database
	if stub.CouchDB {
		return stub.DbHandler.SaveDocument(key, value)
	}
	// Carry on
	return stub.putStateOriginal(key, value)
}

// GetState retrieves the value for a given key from the ledger
func (stub *MockStubExtend) GetState(key string) ([]byte, error) {
	// In case we are using CouchDB, we store the value document in the database
	if stub.CouchDB {
		return stub.DbHandler.ReadDocument(key)
	}
	// Else we can just carry on
	return stub.GetStateOriginal(key)
}

// GetStateOriginal is copied from mockstub as we still need to carry on normal GetState operation with the mock ledger map
func (stub *MockStubExtend) GetStateOriginal(key string) ([]byte, error) {
	value := stub.State[key]
	mockLogger.Debug("MockStub", stub.Name, "Getting", key, value)
	return value, nil
}

// This is copied from mockstub as we still need to carry on normal putstate operation with the mock ledger map
func (stub *MockStubExtend) putStateOriginal(key string, value []byte) error {
	if stub.TxID == "" {
		err := errors.New("cannot PutState without a transactions - call stub.MockTransactionStart()")
		mockLogger.Errorf("%+v", err)
		return err
	}

	// If the value is nil or empty, delete the key
	if len(value) == 0 {
		mockLogger.Debug("MockStub", stub.Name, "PutState called, but value is nil or empty. Delete ", key)
		return stub.DelState(key)
	}

	mockLogger.Debug("MockStub", stub.Name, "Putting", key, value)
	stub.State[key] = value

	// insert key into ordered list of keys
	for elem := stub.Keys.Front(); elem != nil; elem = elem.Next() {
		elemValue := elem.Value.(string)
		comp := strings.Compare(key, elemValue)
		mockLogger.Debug("MockStub", stub.Name, "Compared", key, elemValue, " and got ", comp)
		if comp < 0 {
			// key < elem, insert it before elem
			stub.Keys.InsertBefore(key, elem)
			mockLogger.Debug("MockStub", stub.Name, "Key", key, " inserted before", elem.Value)
			break
		} else if comp == 0 {
			// keys exists, no need to change
			mockLogger.Debug("MockStub", stub.Name, "Key", key, "already in State")
			break
		} else { // comp > 0
			// key > elem, keep looking unless this is the end of the list
			if elem.Next() == nil {
				stub.Keys.PushBack(key)
				mockLogger.Debug("MockStub", stub.Name, "Key", key, "appended")
				break
			}
		}
	}

	// special case for empty Keys list
	if stub.Keys.Len() == 0 {
		stub.Keys.PushFront(key)
		mockLogger.Debug("MockStub", stub.Name, "Key", key, "is first element in list")
	}

	return nil
}

// GetStateByPartialCompositeKey queries couchdb by range
func (stub *MockStubExtend) GetStateByPartialCompositeKey(objectType string, attributes []string) (shim.StateQueryIteratorInterface, error) {
	startKey, _ := stub.CreateCompositeKey(objectType, attributes)
	endKey := startKey + string(maxUnicodeRuneValue)

	rs, er := stub.DbHandler.QueryDocumentByRange(startKey, endKey)
	if er != nil {
		return nil, er
	}

	iterator, er := FromResultsIterator(rs)
	if er != nil {
		return nil, er
	}

	return iterator, nil
}

// GetStateByPartialCompositeKeyWithPagination queries couchdb with a partial compositekey and pagination information
//func (stub *MockStubExtend) GetStateByPartialCompositeKeyWithPagination(objectType string, attributes []string, pageSize int32, bookmark string) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
//	startKey, _ := stub.CreateCompositeKey(objectType, attributes)
//	endKey := startKey + string(maxUnicodeRuneValue)
//
//	rs, er := stub.DbHandler.QueryDocumentByRangeWithPagination(startKey, endKey, pageSize, bookmark)
//	if er != nil{
//		return nil, nil, er
//	}
//
//	iterator, er := FromResultsIterator(rs)
//	if er != nil{
//		return nil, nil, er
//	}
//
//	bm := rs.(statedb.QueryResultsIterator).GetBookmarkAndClose()
//	queryResponse := &pb.QueryResponseMetadata{FetchedRecordsCount: int32(iterator.Length()), Bookmark: bm}
//
//	return iterator, queryResponse, nil
//}
