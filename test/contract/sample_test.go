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

package contract

import (
	"encoding/json"
	"fmt"
	"github.com/Akachain/akc-go-sdk-v2/mock"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"gotest.tools/assert"
	"testing"
)

func setupMock() *mock.MockStubExtend {
	// Initialize MockStubExtend
	sc := new(SampleContract)
	chaincode, _ := contractapi.NewChaincode(sc)
	chaincodeName := "samplecontract"
	stub := mock.NewMockStubExtend(shimtest.NewMockStub(chaincodeName, chaincode), chaincode, ".")

	// Create a new database, Drop old database
	db, err := mock.NewCouchDBHandler(true, chaincodeName)
	if err != nil {
		fmt.Printf("NewCouchDBHandler failed with err (%s)", err.Error())
		return nil
	}
	stub.SetCouchDBConfiguration(db)

	// Process indexes
	err = db.ProcessIndexesForChaincodeDeploy("./META-INF/statedb/couchdb/indexes/indexSampleDoc.json")
	if err != nil {
		return nil
	}
	return stub
}

func TestSimpleData(t *testing.T) {
	stub := setupMock()
	key1 := "key1"
	val1 := "val1"

	mock.MockInvokeTransaction(t, stub, [][]byte{[]byte("CreateSampleObject"), []byte(key1), []byte(val1)})

	// Check if the created data exist in the ledger
	compositeKey, _ := stub.CreateCompositeKey(DocPrefix, []string{key1})
	state, _ := stub.GetState(compositeKey)
	var ad [10]SampleData

	err := json.Unmarshal([]byte(state), &ad[0])
	if err != nil {
		t.Fail()
	}

	// Check if the created data information is correct
	assert.Equal(t, key1, ad[0].Key1)
	assert.Equal(t, val1, ad[0].Attribute1)
}
