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
	util.CheckChaincodeFunctionCallWellFormedness(args, 3)

	approval := new(model.Approval)
	err := json.Unmarshal([]byte(args[0]), approval)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	approval.ApprovalID = stub.GetTxID()

	err = sah.verifySignature(stub, approval.ApproverID, approval.Signature, approval.Message)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	approval.Status = "Verified"

	common.Logger.Infof("Create Approval: %+v\n", approval)
	err = util.CreateData(stub, model.ApprovalTable, []string{approval.ApprovalID}, &approval)
	if err != nil {
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

//GetAllApprovalWithPagination ...
func (sah *ProposalHanler) GetAllApprovalWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var pagesize int32
	errMarshal := json.Unmarshal([]byte(args[0]), &pagesize)
	if errMarshal != nil {
		// Return error: can't unmashal json
		resErr := common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], errMarshal.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}

	res, err := util.QueryAllDataWithPagination(stub, model.ApprovalTable, new(model.Approval), pagesize)
	if err != nil {
		resErr := common.ResponseError{common.ERR3, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}

	fmt.Printf("Datalist: %v\n", res)
	dataJson, err2 := json.Marshal(res)
	if err2 != nil {
		//convert JSON eror
		resErr := common.ResponseError{common.ERR6, common.ResCodeDict[common.ERR6]}
		return common.RespondError(resErr)
	}
	fmt.Printf("Response: %s\n", string(dataJson))
	resSuc := common.ResponseSuccess{common.SUCCESS, common.ResCodeDict[common.SUCCESS], string(dataJson)}
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
