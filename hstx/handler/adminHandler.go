package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/Akachain/akc-go-sdk-v2/common"
	"github.com/Akachain/akc-go-sdk-v2/hstx/model"
	"github.com/Akachain/akc-go-sdk-v2/util"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/mitchellh/mapstructure"
)

// AdminHanler ...
type AdminHanler struct{}

// CreateAdmin ...
func (sah *AdminHanler) CreateAdmin(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 3)

	admin := new(model.Admin)
	err := json.Unmarshal([]byte(args[0]), admin)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	admin.AdminID = stub.GetTxID()
	admin.Status = "Active"

	//common.Logger.Infof("Create Admin: %+v\n", admin)
	log.Printf("Create Admin: %+v\n", admin)
	err = util.CreateData(stub, model.AdminTable, []string{admin.AdminID}, &admin)
	if err != nil {
		resErr := common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		}
		return common.RespondError(resErr)
	}

	bytes, err := json.Marshal(admin)
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

// GetAdminByID ...
func (sah *AdminHanler) GetAdminByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	adminID := args[0]
	res := util.GetDataByIdWithResponse(stub, adminID, new(model.Admin), model.AdminTable)
	return res
}

//UpdateAdmin ...
func (sah *AdminHanler) UpdateAdmin(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	util.CheckChaincodeFunctionCallWellFormedness(args, 1)

	tmpAdmin := new(model.Admin)
	err := json.Unmarshal([]byte(args[0]), tmpAdmin)
	if err != nil {
		// Return error: can't unmashal json
		return common.RespondError(common.ResponseError{
			ResCode: common.ERR3,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine()),
		})
	}

	if len(tmpAdmin.AdminID) == 0 {
		resErr := common.ResponseError{
			ResCode: common.ERR13,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR13], err.Error()),
		}
		return common.RespondError(resErr)
	}

	//get admin information
	rawAdmin, err := util.GetDataById(stub, tmpAdmin.AdminID, model.AdminTable)
	if err != nil {
		resErr := common.ResponseError{
			ResCode: common.ERR4,
			Msg:     fmt.Sprintf("%s %s", common.ResCodeDict[common.ERR4], err.Error()),
		}
		return common.RespondError(resErr)
	}

	admin := new(model.Admin)
	mapstructure.Decode(rawAdmin, admin)

	tmpAdminVal := reflect.ValueOf(tmpAdmin).Elem()
	adminVal := reflect.ValueOf(admin).Elem()
	for i := 0; i < tmpAdminVal.NumField(); i++ {
		fieldName := tmpAdminVal.Type().Field(i).Name
		if len(tmpAdminVal.Field(i).String()) > 0 {
			field := adminVal.FieldByName(fieldName)
			if field.CanSet() {
				field.SetString(tmpAdminVal.Field(i).String())
			}
		}
	}

	err = util.ChangeInfo(stub, model.AdminTable, []string{admin.AdminID}, admin)
	if err != nil {
		//Overwrite fail
		resErr := common.ResponseError{
			ResCode: common.ERR5,
			Msg:     fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR5], err.Error(), common.GetLine()),
		}
		return common.RespondError(resErr)
	}

	bytes, err := json.Marshal(admin)
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
