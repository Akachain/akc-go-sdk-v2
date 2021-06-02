package handler

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Akachain/akc-go-sdk-v2/common"
	"github.com/Akachain/akc-go-sdk-v2/hstx/model"
	"github.com/Akachain/akc-go-sdk-v2/util"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/mitchellh/mapstructure"
)

// ApprovalHanler ...
type ApprovalHanler struct{}

// CreateApproval ...
func (sah *ApprovalHanler) CreateApproval(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check role: SuperAdmin
	err := util.IsSuperAdmin(stub)
	if err != nil {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", err.Error(), common.GetLine()),
		})
	}

	approval := new(model.Approval)
	err = json.Unmarshal([]byte(args[0]), approval)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}
	common.Logger.Debugf("Input-data sent to CreateApproval func: %+v\n", approval)

	// Check SuperAdmin's status
	err = sah.checkApproverStatus(stub, approval.ApproverID)
	if err != nil {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", "This approver is not active", err.Error(), common.GetLine()),
		})
	}

	// Get proposal by approval.ProposalID
	proposalStr, err := new(ProposalHanler).GetProposalByID(stub, approval.ProposalID)
	if err != nil {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine()),
		})
	}

	var proposal model.Proposal
	err = json.Unmarshal([]byte(*proposalStr), &proposal)
	if err != nil {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	// Check whether the proposal was rejected or not
	if strings.Compare("Rejected", proposal.Status) == 0 {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s", "The proposal was rejected", err.Error(), common.GetLine()),
		})
	}

	// Check this approver hasn't signed the proposal
	compositeKey, _ := stub.CreateCompositeKey(model.ApprovalTable, []string{approval.ProposalID, approval.ApproverID})
	rs, err := stub.GetState(compositeKey)
	if err != nil { // Return error: Fail to get data
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine()),
		})
	}
	if len(rs) > 0 { // Return error: Only signing once
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR9,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR9], "This proposal had already been approved", common.GetLine()),
		})
	}

	// Verify signature with the singed message
	err = sah.verifySignature(stub, approval.ApproverID, approval.Signature, approval.Message)
	if err != nil { // Return error: Verify error
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR8,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR8], err.Error(), common.GetLine()),
		})
	}

	// Set approval.ApprovalID & approval.CreatedAt
	approval.ApprovalID = util.GenerateDocumentID(stub)
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine()),
		})
	}
	approval.CreatedAt = time.Unix(timestamp.Seconds, 0).Format(time.RFC3339)

	// Create Approval
	common.Logger.Infof("Creating Approval: %+v\n", approval)
	err = util.CreateData(stub, model.ApprovalTable, []string{approval.ProposalID, approval.ApproverID}, &approval)
	if err != nil { // Return error: Fail to insert data
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		})
	}

	// Update proposal if necessary
	sah.updateProposal(stub, approval)

	bytes, err := json.Marshal(approval)
	if err != nil { // Return error: Can't marshal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	resSuc := common.ResponseSuccess{
		ResCode: common.SUCCESS,
		Msg:     common.ResCodeDict[common.SUCCESS],
		Payload: string(bytes)}
	return common.RespondSuccess(resSuc)
}

// CreateApproval ...
//func (sah *ApprovalHanler) CreateApproval(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	util.CheckChaincodeFunctionCallWellFormedness(args, 3)
//
//	approval := new(model.Approval)
//	err := json.Unmarshal([]byte(args[0]), approval)
//	if err != nil {
//		// Return error: can't unmashal json
//		return common.RespondError(common.ResponseError{
//			ResCode: common.ERR3,
//			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
//		})
//	}
//
//	approval.ApprovalID = stub.GetTxID()
//
//	err = sah.verifySignature(stub, approval.ApproverID, approval.Signature, approval.Message)
//	if err != nil {
//		// Return error: can't unmashal json
//		return common.RespondError(common.ResponseError{
//			ResCode: common.ERR3,
//			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
//		})
//	}
//
//	approval.Status = "Verified"
//
//	common.Logger.Infof("Create Approval: %+v\n", approval)
//	err = util.CreateData(stub, model.ApprovalTable, []string{approval.ApprovalID}, &approval)
//	if err != nil {
//		resErr := common.ResponseError{
//			ResCode: common.ERR5,
//			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
//		}
//		return common.RespondError(resErr)
//	}
//
//	// Update proposal if necessary
//	sah.updateProposal(stub, approval)
//
//	bytes, err := json.Marshal(approval)
//	if err != nil {
//		// Return error: can't mashal json
//		return common.RespondError(common.ResponseError{
//			ResCode: common.ERR5,
//			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
//		})
//	}
//
//	resSuc := common.ResponseSuccess{
//		ResCode: common.SUCCESS,
//		Msg:     common.ResCodeDict[common.SUCCESS],
//		Payload: string(bytes)}
//	return common.RespondSuccess(resSuc)
//}

//GetAllApproval ...
func (sah *ApprovalHanler) GetAllApproval(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	res, err := getApprovalData(stub, 5)
	if err != nil {
		resErr := common.ResponseError{common.ERR3, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}

	result, _ := json.Marshal(res)

	resSuc := common.ResponseSuccess{common.SUCCESS, common.ResCodeDict[common.SUCCESS], string(result)}
	return common.RespondSuccess(resSuc)
}

// GetApprovalByID ...
func (sah *ApprovalHanler) GetApprovalByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	approvalID := args[0]
	res := util.GetDataByIdWithResponse(stub, approvalID, new(model.Approval), model.ApprovalTable)
	return res
}

//UpdateApproval ...
func (sah *ApprovalHanler) UpdateApproval(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	tmpApproval := new(model.Approval)
	err := json.Unmarshal([]byte(args[0]), tmpApproval)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	if len(tmpApproval.ApprovalID) == 0 {
		resErr := common.ResponseError{
			ResCode: common.ERR13,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR13], err.Error()),
		}
		return common.RespondError(resErr)
	}

	//get approval information
	rawApproval, err := util.GetDataById(stub, tmpApproval.ApprovalID, model.ApprovalTable)
	if err != nil {
		resErr := common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR4], err.Error()),
		}
		return common.RespondError(resErr)
	}

	approval := new(model.Approval)
	mapstructure.Decode(rawApproval, approval)

	tmpApprovalVal := reflect.ValueOf(tmpApproval).Elem()
	approvalVal := reflect.ValueOf(approval).Elem()
	for i := 0; i < tmpApprovalVal.NumField(); i++ {
		fieldName := tmpApprovalVal.Type().Field(i).Name
		if len(tmpApprovalVal.Field(i).String()) > 0 {
			field := approvalVal.FieldByName(fieldName)
			if field.CanSet() {
				field.SetString(tmpApprovalVal.Field(i).String())
			}
		}
	}

	err = util.ChangeInfo(stub, model.ApprovalTable, []string{approval.ApprovalID}, approval)
	if err != nil {
		//Overwrite fail
		resErr := common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		}
		return common.RespondError(resErr)
	}

	bytes, err := json.Marshal(approval)
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

// verifySignature ...
func (sah *ApprovalHanler) verifySignature(stub shim.ChaincodeStubInterface, approverID string, signature string, message string) error {

	if len(approverID) == 0 {
		return errors.New("approverID is empty")
	}

	//get superAdmin information
	rawSuperAdmin, err := util.GetDataById(stub, approverID, model.SuperAdminTable)
	if err != nil {
		return err
	}

	superAdmin := new(model.SuperAdmin)
	mapstructure.Decode(rawSuperAdmin, superAdmin)

	// Start verify
	pkBytes := []byte(superAdmin.PublicKey)
	pkBlock, _ := pem.Decode(pkBytes)
	if pkBlock == nil {
		return errors.New("Can't decode public key")
	}

	rawPk, err := x509.ParsePKIXPublicKey(pkBlock.Bytes)
	if err != nil {
		return err
	}

	pk := rawPk.(*ecdsa.PublicKey)

	// SIGNATURE
	signaturebyte, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	R, S, err := utils.UnmarshalECDSASignature(signaturebyte)
	if err != nil {
		return err
	}

	// DATA
	dataByte, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(dataByte)
	var hashData = hash[:]

	// VERIFY
	checksign := ecdsa.Verify(pk, hashData, R, S)

	if checksign {
		return nil
	}
	return errors.New("Verify failed")
}

func getApprovalData(stub shim.ChaincodeStubInterface, pagesize int32) ([]model.Approval, error) {
	//defer lib.TimeTrack(time.Now(), "getTxUsedData", loggerAccountBatch)
	var result = new(model.Approval)
	var list = []model.Approval{}

	var queryString = `
		{ "selector": 
			{
				"_id": 
					{"$gt": "\u0000Approval",
					"$lt": "\u0000Approval\uFFFF"}			
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
		return list, nil
		// return nil, errors.New(lib.ResCodeDict[lib.ERR3])
	}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(queryResponse.Value, result)
		if err != nil {
			continue
		}
		list = append(list, *result)
	}
	return list, nil
}

// updateProposal ...
func (sah *ApprovalHanler) updateProposal(stub shim.ChaincodeStubInterface, approval *model.Approval) error {
	rawProposal, err := util.GetDataById(stub, approval.ProposalID, model.ProposalTable)
	if err != nil {
		return err
	}

	proposal := new(model.Proposal)
	mapstructure.Decode(rawProposal, proposal)

	if strings.Compare(approval.Status, "Rejected") == 0 {
		if strings.Compare(proposal.Status, "Commited") != 0 {
			proposal.Status = approval.Status
			proposal.UpdatedAt = approval.CreatedAt
			bytes, err := json.Marshal(proposal)
			if err != nil {
				return err
			}
			new(ProposalHanler).UpdateProposal(stub, []string{string(bytes)})
		}
		return nil
	}

	resIterator, err := util.GetContainKey(stub, model.ApprovalTable, approval.ProposalID)
	if err != nil {
		return err
	}
	defer resIterator.Close()
	count := 0
	if approval.Status == "Approved" {
		count++
	}
	for resIterator.HasNext() {
		stateIterator, err := resIterator.Next()
		if err != nil {
			return err
		}
		approvalState := new(model.Approval)
		err = json.Unmarshal(stateIterator.Value, approvalState)
		if err != nil { // Convert JSON error
			return err
		}

		if strings.Compare("Approved", approvalState.Status) == 0 {
			count++
		}
	}
	// Check approved number >= proposal.QuorumNumber to update the Proposal's satatus
	if count >= proposal.QuorumNumber {
		rawProposal, err := util.GetDataById(stub, approval.ProposalID, model.ProposalTable)
		if err != nil {
			return err
		}
		proposal := new(model.Proposal)
		mapstructure.Decode(rawProposal, proposal)
		if strings.Compare(proposal.Status, "Pending") == 0 {
			proposal.Status = "Approved"
			proposal.UpdatedAt = approval.CreatedAt
			bytes, err := json.Marshal(proposal)
			if err != nil {
				return err
			}
			new(ProposalHanler).UpdateProposal(stub, []string{string(bytes)})
		}
	}
	return nil
}

// checkApproverStatus func to check whether the SuperAdmin is active or inactive
func (sah *ApprovalHanler) checkApproverStatus(stub shim.ChaincodeStubInterface, approverID string) error {
	// Get approver by approval.ApproverID
	superAdminStr := new(SuperAdminHanler).GetSuperAdminByID(stub, []string{approverID})
	//if err != nil {
	//	return fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())
	//}

	sAdmin := superAdminStr.Payload

	var superAdmin model.SuperAdmin
	err := json.Unmarshal([]byte(sAdmin), &superAdmin)
	if err != nil {
		return fmt.Errorf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())
	}

	// Check SuperAdmin's status
	if superAdmin.Status != "A" && superAdmin.Status != "Active" {
		return fmt.Errorf("%s %s", "This approver is not active", common.GetLine())
	}
	// If the SuperAdmin is active, return nil
	return nil
}
