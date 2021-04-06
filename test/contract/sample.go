package contract

import (
	"github.com/Akachain/akc-go-sdk-v2/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SampleContract struct {
	contractapi.Contract
}

// Prefix for key in stateDB
const DocPrefix = "SAMPLE"

// Data - struct
type SampleData struct {
	Key1       string `json:"Key1"`
	Attribute1 string `json:"Attribute1"`
}

// Create something
func (s *SampleContract) CreateSampleObject(ctx contractapi.TransactionContextInterface, key string, val string) error {
	return util.Createdata(ctx.GetStub(), DocPrefix, []string{key}, &SampleData{Key1: key, Attribute1: val})
}