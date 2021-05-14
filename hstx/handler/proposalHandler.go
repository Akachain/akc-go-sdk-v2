package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Akachain/akc-go-sdk-v2/common"
	"github.com/Akachain/akc-go-sdk-v2/hstx/model"
	"github.com/Akachain/akc-go-sdk-v2/util"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/mitchellh/mapstructure"
)

// ProposalHanler ...
type ProposalHanler struct{}

// CreateProposal ...
func (sah *ProposalHanler) CreateProposal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	common.Logger.Infof("Create Proposal: %+v\n", args)

	proposal := new(model.Proposal)
	err := json.Unmarshal([]byte(args[0]), proposal)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	proposal.ProposalID = stub.GetTxID()
	proposal.Status = "Pending"

	common.Logger.Infof("Create Proposal: %+v\n", proposal)
	err = util.CreateData(stub, model.ProposalTable, []string{proposal.ProposalID}, &proposal)
	if err != nil {
		resErr := common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		}
		return common.RespondError(resErr)
	}

	bytes, err := json.Marshal(proposal)
	if err != nil {
		// Return error: can't mashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		})
	}

	resSuc := common.ResponseSuccess{
		ResCode: common.SUCCESS,
		Msg:     common.ResCodeDict[common.SUCCESS],
		Payload: string(bytes)}
	return common.RespondSuccess(resSuc)
}

//GetAllProposal ...
func (sah *ProposalHanler) GetAllProposal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//var pagesize int32
	//errMarshal := json.Unmarshal([]byte(args[0]), &pagesize)
	//if errMarshal != nil {
	//	// Return error: can't unmashal json
	//	resErr := common.ResponseError{
	//		ResCode: common.ERR4,
	//		Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], errMarshal.Error(), common.GetLine())}
	//	return common.RespondError(resErr)
	//}

	res, err := util.QueryAllDataWithPagination(stub, model.ProposalTable, new(model.Proposal), 5)
	//res, err := getProposalData(stub, 5)
	if err != nil {
		resErr := common.ResponseError{common.ERR3, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}

	result, _ := json.Marshal(res)

	resSuc := common.ResponseSuccess{common.SUCCESS, common.ResCodeDict[common.SUCCESS], string(result)}
	return common.RespondSuccess(resSuc)
}

// GetProposalByID ...
func (sah *ProposalHanler) GetProposalByID(stub shim.ChaincodeStubInterface, proposalID string) (result *string, err error) {
	common.Logger.Debugf("Input-data sent to GetProposalByID func: %+v\n", proposalID)

	rawProposal, err := util.GetDataById(stub, proposalID, model.ProposalTable)
	if err != nil {
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
	}

	proposal := new(model.Proposal)
	mapstructure.Decode(rawProposal, proposal)

	bytes, err := json.Marshal(proposal)
	if err != nil { // Return error: Can't marshal json
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
	}
	temp := ""
	result = &temp
	*result = string(bytes)

	return result, nil
}

// GetPendingProposalBySuperAdminID ...
func (sah *ProposalHanler) GetPendingProposalBySuperAdminID(stub shim.ChaincodeStubInterface, superAdminID string) (result *string, err error) {
	common.Logger.Debugf("Input-data sent to GetPendingProposalBySuperAdminID func: %+v\n", superAdminID)

	var proposalList []model.Proposal

	queryStr := fmt.Sprintf("{\"selector\": {\"_id\": {\"$regex\": \"%s\"},\"$or\": [{\"Status\": \"Pending\"},{\"Status\": \"Approved\"}]}}", model.ProposalTable)
	resultsIterator, err := stub.GetQueryResult(queryStr)
	if err != nil {
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
		}

		proposal := new(model.Proposal)
		err = json.Unmarshal(queryResponse.Value, proposal)
		if err != nil { // Convert JSON error
			return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
		}
		proposalList = append(proposalList, *proposal)
	}

	for i := len(proposalList) - 1; i >= 0; i-- {
		proposal := proposalList[i]
		rs, err := util.GetByTwoColumns(stub, model.ApprovalTable, "ProposalID", fmt.Sprintf("\"%s\"", proposal.ProposalID), "ApproverID", fmt.Sprintf("\"%s\"", superAdminID))
		if err != nil {
			return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
		}
		if rs.HasNext() {
			proposalList[i] = proposalList[len(proposalList)-1]
			proposalList = proposalList[:len(proposalList)-1]
		}
	}

	bytes, err := json.Marshal(proposalList)
	if err != nil { // Return error: Can't marshal json
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
	}
	temp := ""
	result = &temp
	*result = string(bytes)

	return result, nil
}

//UpdateProposal ...
func (sah *ProposalHanler) UpdateProposal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	tmpProposal := new(model.Proposal)
	err := json.Unmarshal([]byte(args[0]), tmpProposal)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	if len(tmpProposal.ProposalID) == 0 {
		resErr := common.ResponseError{
			ResCode: common.ERR13,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR13], err.Error()),
		}
		return common.RespondError(resErr)
	}

	//get proposal information
	rawProposal, err := util.GetDataById(stub, tmpProposal.ProposalID, model.ProposalTable)
	if err != nil {
		resErr := common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR4], err.Error()),
		}
		return common.RespondError(resErr)
	}

	proposal := new(model.Proposal)
	mapstructure.Decode(rawProposal, proposal)

	tmpProposalVal := reflect.ValueOf(tmpProposal).Elem()
	proposalVal := reflect.ValueOf(proposal).Elem()
	for i := 0; i < tmpProposalVal.NumField(); i++ {
		fieldName := tmpProposalVal.Type().Field(i).Name
		if len(tmpProposalVal.Field(i).String()) > 0 {
			field := proposalVal.FieldByName(fieldName)
			if field.CanSet() {
				field.SetString(tmpProposalVal.Field(i).String())
			}
		}
	}

	err = util.ChangeInfo(stub, model.ProposalTable, []string{proposal.ProposalID}, proposal)
	if err != nil {
		//Overwrite fail
		resErr := common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		}
		return common.RespondError(resErr)
	}

	bytes, err := json.Marshal(proposal)
	if err != nil {
		// Return error: can't mashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		})
	}

	resSuc := common.ResponseSuccess{
		ResCode: common.SUCCESS,
		Msg:     common.ResCodeDict[common.SUCCESS],
		Payload: string(bytes)}
	return common.RespondSuccess(resSuc)
}

//CommitProposal ...
func (sah *ProposalHanler) CommitProposal(stub shim.ChaincodeStubInterface, proposalID string) (result *string, err error) {
	common.Logger.Debugf("Input-data sent to CommitProposal func: %+v\n", proposalID)

	proposalStr, err := sah.GetProposalByID(stub, proposalID)
	if err != nil {
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
	}

	var proposal model.Proposal
	err = json.Unmarshal([]byte(*proposalStr), &proposal)
	if err != nil {
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
	}

	if strings.Compare("Pending", proposal.Status) == 0 {
		return nil, fmt.Errorf("%s %s", "Not enough approval", common.GetLine())
	}

	if strings.Compare("Rejected", proposal.Status) == 0 {
		return nil, fmt.Errorf("%s %s", "The proposal was rejected", common.GetLine())
	}

	proposal.Status = "Committed"
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
	}
	updatedTime := time.Unix(timestamp.Seconds, 0)
	proposal.UpdatedAt = updatedTime.String()

	err = util.ChangeInfo(stub, model.ProposalTable, []string{proposal.ProposalID}, proposal)
	if err != nil { // Return error: Fail to Update data
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine())
	}

	bytes, err := json.Marshal(proposal)
	if err != nil { // Return error: Can't marshal json
		return nil, fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
	}
	temp := ""
	result = &temp
	*result = string(bytes)

	return result, nil
}

func getProposalData(stub shim.ChaincodeStubInterface, pagesize int32) ([]model.Proposal, error) {
	//defer lib.TimeTrack(time.Now(), "getTxUsedData", loggerAccountBatch)
	var txUsedResult = new(model.Proposal)
	var txUsedList = []model.Proposal{}

	var queryString = `
		{ "selector": 
			{
				"_id": 
					{"$gt": "\u0000Proposal",
					"$lt": "\u0000Proposal\uFFFF"}			
			}
		}`

	common.Logger.Debugf("Get Query String %s", queryString)
	resultsIterator, _, err := stub.GetQueryResultWithPagination(queryString, pagesize, "")
	common.Logger.Debug("Finish Get query")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// Check data respose after query in database
	if !resultsIterator.HasNext() {
		// Return with txUsedList empty
		return txUsedList, nil
		// return nil, errors.New(lib.ResCodeDict[lib.ERR3])
	}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(queryResponse.Value, txUsedResult)
		if err != nil {
			continue
		}
		txUsedList = append(txUsedList, *txUsedResult)
	}
	return txUsedList, nil
}
