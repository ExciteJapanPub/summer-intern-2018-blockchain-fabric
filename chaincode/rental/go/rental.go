/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type EquipmentData struct {
	EquipmentId   string        `json:"equipment_id"`   // "0001"
	EquipmentName string        `json:"equipment_name"` // "scissors"
	Total         int           `json:"total"`          // 10
	BorrowerList  []string      `json:"borrower_list"`  // ["1726_takarada"]
}

type UserData struct {
	UserId      string        `json:"user_id"`      // "1726_takarada"
	ReturnDate  string        `json:"return_date"`  // "2018/08/22"
	IsBorrowing bool          `json:"is_borrowing"` // true
	EquipmentId string        `json:"equipment_id"` // "0001"
}

type Status int

const (
	StatusOk Status = 200
	StatusNotFound Status = 404
	StatusNotAllowed Status = 405
	StatusConflict Status = 409
)

type ErrorResult struct {
	Status  Status        `json:"status"`
	Message string        `json:"message"`
}

type GetUserDataResult struct {
	Status   Status        `json:"status"`
	UserData UserData      `json:"user_data"`
}
type RegisterUserResult GetUserDataResult

type GetEquipmentDataResult struct {
	Status        Status           `json:"status"`
	EquipmentData EquipmentData    `json:"equipment_data"`
}

type AddEquipmentResult GetEquipmentDataResult
type BorrowEquipmentResult GetEquipmentDataResult
type ReturnEquipmentResult GetEquipmentDataResult

const DateFormat = "2006/01/02"
const DefaultUserIsBorrowing bool = false
const AdminUserId string = "0001_admin"

// Define the Smart Contract structure
type SmartContract struct {
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "getUserData" {
		return s.getUserData(APIstub, args)
	}
	if function == "registerUserData" {
		return s.registerUserData(APIstub, args)
	}
	if function == "getEquipmentData" {
		return s.getEquipmentData(APIstub, args)
	}
	if function == "registerEquipmentData" {
		return s.registerEquipmentData(APIstub, args)
	}
	if function == "borrowEquipment" {
		return s.borrowEquipment(APIstub, args)
	}
	if function == "returnEquipment" {
		return s.returnEquipment(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return s.initAdminUser(APIstub)
}

// adminユーザデータの初期化
func (s *SmartContract) initAdminUser(APIstub shim.ChaincodeStubInterface) sc.Response {
	adminUser := s.getUserDataFromState(APIstub, AdminUserId)
	// すでに有効なadminユーザがあるならそれを返却
	if adminUser.UserId != "" {
		result := GetUserDataResult{Status: StatusOk, UserData: adminUser}
		resultAsBytes, _ := json.Marshal(result)
		return shim.Success(resultAsBytes)
	}

	// 有効なadminユーザがなければ作成してput
	adminUser.UserId = AdminUserId
	adminUser.IsBorrowing = DefaultUserIsBorrowing
	s.putUserData(APIstub, adminUser)

	result := GetUserDataResult{Status: StatusOk, UserData: adminUser}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// エラーレスポンスを生成する
func (s *SmartContract) makeErrorResponse(APIstub shim.ChaincodeStubInterface, code Status, message string) sc.Response {
	result := ErrorResult{
		Status:  code,
		Message: message,
	}

	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// UserData用Stateキー作成関数
func (s *SmartContract) makeUserDataKey(userId string) string {
	return "user_" + userId
}

// equipment用Stateキー作成関数
func (s *SmartContract) makeEquipmentDataKey(equipmentId string) string {
	return "equipment_" + equipmentId
}

// 指定備品のデータ取得
func (s *SmartContract) getEquipmentDataFromState(APIstub shim.ChaincodeStubInterface, equipmentId string) EquipmentData {
	key := s.makeEquipmentDataKey(equipmentId)

	equipmentDataAsBytes, _ := APIstub.GetState(key)
	equipment := new(EquipmentData)

	if len(equipmentDataAsBytes) != 0 {
		json.Unmarshal(equipmentDataAsBytes, equipment)
	}

	return *equipment
}

// 指定ユーザーのデータ取得
func (s *SmartContract) getUserDataFromState(APIstub shim.ChaincodeStubInterface, userId string) UserData {
	key := s.makeUserDataKey(userId)

	userDataAsBytes, _ := APIstub.GetState(key)
	userData := new(UserData)

	if len(userDataAsBytes) != 0 {
		json.Unmarshal(userDataAsBytes, userData)
	}

	return *userData
}

// ユーザーデータput
func (s *SmartContract) putUserData(APIstub shim.ChaincodeStubInterface, userData UserData) {
	key := s.makeUserDataKey(userData.UserId)
	userDataAsBytes, _ := json.Marshal(userData)
	APIstub.PutState(key, userDataAsBytes)
}

// 備品put
func (s *SmartContract) putEquipmentData(APIstub shim.ChaincodeStubInterface, equipment EquipmentData) {
	key := s.makeEquipmentDataKey(equipment.EquipmentId)
	equipmentDataAsBytes, _ := json.Marshal(equipment)
	APIstub.PutState(key, equipmentDataAsBytes)
}

// 配列の中身削除
func remove(strings []string, search string) []string {
	for i, v := range strings {
		if v == search {
			return append(strings[0:i], strings[i+1:len(strings)]...)
		}
	}
	return strings
}

// ユーザーデータの取得
func (s *SmartContract) getUserData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userId := args[0]

	userData := s.getUserDataFromState(APIstub, userId)
	if userData.UserId == "" {
		return s.makeErrorResponse(APIstub, StatusNotFound, "対象のユーザIDのデータが存在しません")
	}

	result := GetUserDataResult{Status: StatusOk, UserData: userData}
	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

// ユーザーデータの登録
func (s *SmartContract) registerUserData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userId := args[0]

	userData := s.getUserDataFromState(APIstub, userId)
	// すでにユーザーデータが存在するなら登録不可
	if userData.UserId != "" {
		return s.makeErrorResponse(APIstub, StatusConflict, "対象のユーザーIDのデータが存在しています")
	}

	// ユーザーデータがなければ空のデータを作成してput
	userData.UserId = userId
	userData.IsBorrowing = DefaultUserIsBorrowing
	s.putUserData(APIstub, userData)

	result := RegisterUserResult{Status: StatusOk, UserData: userData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// 備品データの取得
func (s *SmartContract) getEquipmentData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	key := s.makeEquipmentDataKey(args[0])
	equipmentDataAsBytes, _ := APIstub.GetState(key)
	equipmentData := EquipmentData{}
	json.Unmarshal(equipmentDataAsBytes, &equipmentData)

	result := GetEquipmentDataResult{Status: StatusOk, EquipmentData: equipmentData}
	if equipmentData.EquipmentId == "" {
		return s.makeErrorResponse(APIstub, StatusNotFound, "対象の備品IDのデータが存在しません")
	}

	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

// 備品データの追加
func (s *SmartContract) registerEquipmentData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	userId := args[0]
	equipmentId := args[1]
	equipmentName := args[2]
	equipmentTotal, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Incorrect argument. fourth argument is should be numerical string")
	}

	// adminユーザでなければ備品を追加できない
	if userId != AdminUserId {
		return s.makeErrorResponse(APIstub, StatusNotAllowed, "adminユーザーでなければ備品を追加できません")
	}

	equipmentData := s.getEquipmentDataFromState(APIstub, equipmentId)
	// すでに備品データが存在するなら追加不可
	if equipmentData.EquipmentId != "" {
		return s.makeErrorResponse(APIstub, StatusConflict, "対象の備品IDのデータが存在しています")
	}

	// 備品データがなければ新しく備品追加
	equipmentData.EquipmentId = equipmentId
	equipmentData.EquipmentName = equipmentName
	equipmentData.Total = equipmentTotal

	s.putEquipmentData(APIstub, equipmentData)

	result := AddEquipmentResult{Status: StatusOk, EquipmentData: equipmentData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// 備品を借りる
func (s *SmartContract) borrowEquipment(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	userId := args[0]
	equipmentId := args[1]
	equipmentReturnDate := args[2]

	userData := s.getUserDataFromState(APIstub, userId)
	// ユーザーデータが存在しないなら借りられない
	if userData.UserId == "" {
		return s.makeErrorResponse(APIstub, StatusNotFound, "対象のユーザーIDのデータが存在しません")
	}

	// ユーザーがすでに備品を借りているなら新しく備品は借りられない
	if userData.IsBorrowing == true {
		return s.makeErrorResponse(APIstub, StatusNotAllowed, "すでに備品を借りています。返却してから再度借り直してください。")
	}

	equipmentData := s.getEquipmentDataFromState(APIstub, equipmentId)
	// 備品データが存在しないなら借りられない
	if equipmentData.EquipmentId == "" {
		return s.makeErrorResponse(APIstub, StatusNotFound, "対象の備品IDのデータが存在しません")
	}

	// 備品の残りがないなら借りられない
	stock := equipmentData.Total - len(equipmentData.BorrowerList)
	if stock == 0 {
		return s.makeErrorResponse(APIstub, StatusNotFound, "備品の在庫がありません")
	}

	// 借りる処理
	returnDate, err := time.Parse(DateFormat, equipmentReturnDate)
	if err != nil {
		return shim.Error("Format of date should be yyyy/mm/dd")
	}
	userData.ReturnDate = returnDate.Format(DateFormat)
	userData.IsBorrowing = true
	userData.EquipmentId = equipmentId
	equipmentData.BorrowerList = append(equipmentData.BorrowerList, userId)

	s.putUserData(APIstub, userData)
	s.putEquipmentData(APIstub, equipmentData)

	result := BorrowEquipmentResult{Status: StatusOk, EquipmentData: equipmentData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// 備品を返却
func (s *SmartContract) returnEquipment(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userId := args[0]

	userData := s.getUserDataFromState(APIstub, userId)
	// ユーザーデータが存在しない
	if userData.UserId == "" {
		return s.makeErrorResponse(APIstub, StatusNotFound, "対象のユーザーIDのデータが存在しません")
	}

	// ユーザーが備品を借りてない
	if userData.IsBorrowing == false {
		return s.makeErrorResponse(APIstub, StatusNotAllowed, "いずれの備品も借りていません。")
	}

	equipmentData := s.getEquipmentDataFromState(APIstub, userData.EquipmentId)

	userData.ReturnDate = ""
	userData.IsBorrowing = false
	userData.EquipmentId = ""
	equipmentData.BorrowerList = remove(equipmentData.BorrowerList, userId)

	s.putUserData(APIstub, userData)
	s.putEquipmentData(APIstub, equipmentData)

	result := ReturnEquipmentResult{Status: StatusOk, EquipmentData: equipmentData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
