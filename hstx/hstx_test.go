package main

import (
	"fmt"
	"testing"

	"github.com/Akachain/akc-go-sdk/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkCallFunc(t *testing.T, stub *util.MockStubExtend, args [][]byte) string {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		return string(res.Payload)
	}
	return string(res.Payload)
}

func TestMainchain(t *testing.T) {
	cc := new(Chaincode)
	mock := shim.NewMockStub("hstx", cc)
	stub := util.NewMockStubExtend(mock)

	AdminID := "ntienbo"

	// Check get CreateProposal
	// rs := checkCallFunc(t, stub, [][]byte{[]byte("CreateProposal"), []byte("1")})

	// rs5 := checkCallFunc(t, stub, [][]byte{[]byte("CreateAdmin"), []byte("Admin"), []byte("pulic key")})
	// rs7 := checkCallFunc(t, stub, [][]byte{[]byte("CreateAdmin"), []byte("Admin1"), []byte("pulic key")})

	//rs3 := checkByID(t, stub, [][]byte{[]byte("GetAdminByID1"), []byte("aa"), []byte(models.ADMINTABLE)})

	//rs4 := checkGetALL(t, stub, "GetAllProposal")

	// rs6 := checkCallFunc(t, stub, [][]byte{[]byte("GetAllData"), []byte(models.ADMINTABLE)})

	ProposalID := "d"
	sig := "MEYCIQC7vKLzjw43HJ/9SqxUzZtfdBIdFks7qiIXiHitu8uqqQIhAKXNwpBDuWquPE/00l8isa6rh85ZYYf+dgb1khSqNr7O"
	rs2 := checkCallFunc(t, stub, [][]byte{[]byte("CreateQuorum"), []byte(sig), []byte(AdminID), []byte(ProposalID)})

	// fmt.Printf("rs: %v", rs)
	fmt.Printf("rs2: %v", rs2)
	// fmt.Printf("rs3: %v", rs3)
	//	fmt.Printf("rs4: %v", rs4)
	// fmt.Printf("rs5: %v", rs5)
	// fmt.Printf("rs6: %v", rs6)
	// fmt.Printf("rs6: %v", rs7)

	// check checkByID
}