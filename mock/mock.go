package mock

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/satori/go.uuid"
	"testing"
)

// MockInvokeTransaction creates a mock invoke transaction using MockStubExtend
func MockInvokeTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInvoke(txId, args)
	if res.Status != shim.OK {
		return string(res.Message)
	}
	// fmt.Println(res.Payload)
	return string(res.Payload)
}

// MockQueryTransaction creates a mock query transaction using MockStubExtend
func MockQueryTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInvoke(txId, args)
	if res.Status != shim.OK {
		t.FailNow()
		return string(res.Message)
	}
	return string(res.Payload)
}

// MockIInit creates a mock invoke transaction using MockStubExtend
func MockInitTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInit(txId, args)
	if res.Status != shim.OK {
		return string(res.Message)
	}
	return string(res.Payload)
}

// Generate random transaction ID
func genTxID() string {
	// or error handling
	uid := uuid.NewV4()
	txId := fmt.Sprintf("%s", uid)
	return txId
}