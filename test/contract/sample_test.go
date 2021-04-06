package contract

import (
	"encoding/json"
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
	db, _ := mock.NewCouchDBHandler(true, chaincodeName)
	stub.SetCouchDBConfiguration(db)
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
	var ad [10] SampleData

	json.Unmarshal([]byte(state), &ad[0])

	// Check if the created data information is correct
	assert.Equal(t, key1, ad[0].Key1)
	assert.Equal(t, val1, ad[0].Attribute1)
}