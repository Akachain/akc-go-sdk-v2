package common

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
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
		Payload: []byte("{\"status\":\"" + res.ResCode + "\", \"msg\":\"" + res.Msg + "\", \"rows\":" + res.Payload + "}"),
	}
}

func RespondError(err ResponseError) pb.Response {
	msg := "{\"status\":\"" + err.ResCode + "\", \"msg\":\"" + err.Msg + "\"}"
	return pb.Response{
		Status:  ERROR,
		Message: msg,
	}
}

func GetQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}