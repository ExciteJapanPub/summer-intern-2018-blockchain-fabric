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
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type UserData struct {
	UserId                   string        `json:"user_id"`                      // "kohun_0001"
	CardIdmHash              string        `json:"card_idm_hash"`                // "gq8235u8qweguba1se3bose6irfsd"
	LastChangeStatusLockerId string        `json:"last_change_status_locker_id"` // "box_0001"
}

type LockerData struct {
	LockerId               string        `json:"locker_id"`                  // "box_0001"
	LockerStatus           LockerStatus  `json:"from_user_id"`               //"locked"
	LastChangeStatusTime   string        `json:"last_change_status_time"`    // "2018-09-22 20:32:11 UTC"
	LastChangeStatusUserId string        `json:"last_change_status_user_id"` // "kohun_0001"
	AllowedUnlockUserIds   []string      `json:"allowed_unlock_user_ids"`    // ["kohun_0001"]
}

type Status int

const (
	StatusOk Status = 200
	StatusNotFound Status = 404
	StatusNotAllowed Status = 405
	StatusConflict Status = 409
)

type LockerStatus string

const (
	StatusLock LockerStatus = "locked"
	StatusUnlock LockerStatus = "unlocked"
)

const DateTimeFormat = "2006-01-02 15:04:05 UTC"
const DefaultLockerId string = "box_0001"

type GetUserDataResult struct {
	Status   Status        `json:"status"`
	UserData UserData      `json:"user_data"`
}
type RegisterUserResult GetUserDataResult

type GiveLockerPermissionResult struct {
	Status Status        `json:"status"`
	UserId string        `json:"user_id"`
}

type GetLockerDataResult struct {
	Status     Status        `json:"status"`
	LockerData LockerData    `json:"locker_data"`
}
type ChangeLockerStatusResult GetLockerDataResult

type ErrorResult struct {
	Status  Status        `json:"status"`
	Message string        `json:"message"`
}

// Define the Smart Contract structure
type SmartContract struct {
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return s.initDefaultLocker(APIstub)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "getUserData" {
		return s.getUserData(APIstub, args)
	}
	if function == "getLockerData" {
		return s.getLockerData(APIstub, args)
	}
	if function == "registerUser" {
		return s.registerUser(APIstub, args)
	}
	if function == "giveLockerPermission" {
		return s.giveLockerPermission(APIstub, args)
	}
	if function == "changeLockerStatus" {
		return s.changeLockerStatus(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// UserData用Stateキー作成関数
func (s *SmartContract) makeUserDataKey(userId string) string {
	return "user_" + userId
}

// lockerData用Stateキー作成関数
func (s *SmartContract) makeLockerDataKey(lockerId string) string {
	return "locker_" + lockerId
}

// 空のLockerData構造体作成
func (s *SmartContract) makeEmptyLockerData(lockerId string) LockerData {
	result := LockerData{
		LockerId: lockerId,
		LockerStatus: StatusLock,
		LastChangeStatusTime:  time.Now().Format(DateTimeFormat),
		LastChangeStatusUserId:  "",
		AllowedUnlockUserIds: []string{},
	}
	return result
}

// エラーレスポンスを生成する
func (s *SmartContract) makeErrorResponce(APIstub shim.ChaincodeStubInterface, code Status, message string) sc.Response {
	result := ErrorResult{
		Status:  code,
		Message: message,
	}

	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
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

// 指定ロッカーのデータ取得
func (s *SmartContract) getLockerDataFromState(APIstub shim.ChaincodeStubInterface, lockerId string) LockerData {
	key := s.makeLockerDataKey(lockerId)

	lockerDataAsBytes, _ := APIstub.GetState(key)
	lockerData := new(LockerData)

	if len(lockerDataAsBytes) != 0 {
		json.Unmarshal(lockerDataAsBytes, lockerData)
	}

	return *lockerData
}

// ロッカーデータが有効であるか判別する
func (s *SmartContract) isValidLockerData(lockerData LockerData) bool {
	if lockerData.LockerId == "" {
		return false
	}

	// 論理削除や有効期限など追加した場合はここでチェックする

	return true
}

// ユーザーデータが有効であるか判別する
func (s *SmartContract) isValidUserData(userData UserData) bool {
	if userData.CardIdmHash == "" {
		return false
	}

	// 論理削除や有効期限など追加した場合はここでチェックする

	return true
}

// ロッカーデータのステータスが変更可能かどうか
func (s *SmartContract) canChangeLockerStatus(lockerData LockerData, toStatus LockerStatus) bool {
	if lockerData.LockerStatus == toStatus {
		return false
	}

	// 開錠のステータス変更条件が他にあればここでチェックする

	return true
}

// ロッカーを開錠できるかどうか判別する
func (s *SmartContract) isAuthorizedUserForLocker(userData UserData, lockerData LockerData) bool {

	for _, authorizedUserId := range lockerData.AllowedUnlockUserIds {
		if authorizedUserId == userData.UserId {
			return true
		}
	}

	return false
}

// ユーザーデータput
func (s *SmartContract) putUserData(APIstub shim.ChaincodeStubInterface, userData UserData) {
	key := s.makeUserDataKey(userData.UserId)
	userDataAsBytes, _ := json.Marshal(userData)
	APIstub.PutState(key, userDataAsBytes)
}

// ロッカーデータput
func (s *SmartContract) putLockerData(APIstub shim.ChaincodeStubInterface, lockerData LockerData) {
	key := s.makeLockerDataKey(lockerData.LockerId)
	lockerDataAsBytes, _ := json.Marshal(lockerData)
	APIstub.PutState(key, lockerDataAsBytes)
}

// デフォルトのロッカーデータの初期化
func (s *SmartContract) initDefaultLocker(APIstub shim.ChaincodeStubInterface) sc.Response {
	defaultLocker := s.getLockerDataFromState(APIstub, DefaultLockerId)
	// すでに有効なロッカーデータがあるならそれを返却
	if s.isValidLockerData(defaultLocker) {
		result := GetLockerDataResult{Status: StatusOk, LockerData: defaultLocker}
		resultAsBytes, _ := json.Marshal(result)
		return shim.Success(resultAsBytes)
	}

	// 有効なロッカーがなければ空のデータを作成してput
	emptyLockerData := s.makeEmptyLockerData(DefaultLockerId)
	s.putLockerData(APIstub, emptyLockerData)

	result := GetLockerDataResult{Status: StatusOk, LockerData: emptyLockerData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// ユーザーデータの登録
func (s *SmartContract) registerUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	userId := args[0]
	cardIdmHash := args[1]

	userData := s.getUserDataFromState(APIstub, userId)
	// すでに有効なユーザーデータがあるなら登録不可
	if s.isValidUserData(userData) {
		return s.makeErrorResponce(APIstub, StatusConflict, "対象のユーザーIDのデータが存在しています")
	}

	// ユーザーデータがなければ空のデータを作成してput
	userData.UserId = userId
	userData.CardIdmHash = cardIdmHash
	s.putUserData(APIstub, userData)

	result := RegisterUserResult{Status: StatusOk, UserData: userData}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// ロッカーに対して操作可能なユーザーを追加する
func (s *SmartContract) giveLockerPermission(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	userId := args[0]
	lockerId := args[1]

	lockerData := s.getLockerDataFromState(APIstub, lockerId)

	//追加の制限かけるならここで色々チェック


	lockerData.AllowedUnlockUserIds = append(lockerData.AllowedUnlockUserIds, userId)
	s.putLockerData(APIstub, lockerData)

	result := GiveLockerPermissionResult{Status: StatusOk, UserId: userId}
	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getUserData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := s.makeUserDataKey(args[0])
	userDataAsBytes, _ := APIstub.GetState(key)
	userData := UserData{}
	json.Unmarshal(userDataAsBytes, &userData)

	result := GetUserDataResult{Status: StatusOk, UserData: userData}
	if userData.CardIdmHash == "" {
		result.Status = StatusNotFound
	}

	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getLockerData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	key := s.makeLockerDataKey(args[0])
	lockerDataAsBytes, _ := APIstub.GetState(key)
	lockerData := LockerData{}
	json.Unmarshal(lockerDataAsBytes, &lockerData)

	result := GetLockerDataResult{Status: StatusOk, LockerData: lockerData}
	if lockerData.LockerId == "" {
		result.Status = StatusNotFound
	}

	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

func (s *SmartContract) changeLockerStatus(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	userId := args[0]
	lockerId := args[1]
	toStatus := LockerStatus(args[2])

	userData := s.getUserDataFromState(APIstub, userId)
	lockerData := s.getLockerDataFromState(APIstub, lockerId)

	//対象のロッカーの開錠権限があるかどうか
	if !s.isAuthorizedUserForLocker(userData, lockerData) {
		return s.makeErrorResponce(APIstub, StatusNotAllowed, "対象のロッカーの開錠権限がありません")
	}

	//ロッカーの開閉ができるかどうか
	if !s.canChangeLockerStatus(lockerData, toStatus) {
		return s.makeErrorResponce(APIstub, StatusConflict, "対象のロッカーの開閉ができませんでした")
	}

	lockerData.LockerStatus = toStatus
	lockerData.LastChangeStatusUserId = userData.UserId
	lockerData.LastChangeStatusTime = time.Now().Format(DateTimeFormat)
	userData.LastChangeStatusLockerId = lockerData.LockerId

	s.putLockerData(APIstub, lockerData)
	s.putUserData(APIstub, userData)

	// レスポンス作成
	result := ChangeLockerStatusResult{
		Status:     StatusOk,
		LockerData: lockerData,
	}
	resultAsBytes, _ := json.Marshal(result)

	// 返却
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
