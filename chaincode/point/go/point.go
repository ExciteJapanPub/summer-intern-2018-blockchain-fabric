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

type Balance struct {
	UserId string  `json:"user_id"` // "1725"
	Amount float32 `json:"amount"`  // 2000.000
	Total  float32 `json:"total"`   // 2500.000
}

type TransferHistory struct {
	ToUserId   string  `json:"to_user_id"`   // "1725"
	FromUserId string  `json:"from_user_id"` // "0001"
	Point      float32 `json:"price"`        // 300.000
}

type TransferHistoryWithTimestamp struct {
	TransferHistory
	CreatedAt string `json:"created_at"` // "2018-07-01 20:32:11 UTC"
}

type Status int

const (
	StatusOk         Status = 200
	StatusBadRequest Status = 400
	StatusNotFound   Status = 404
)

const MonthFormat = "200601"
const DateTimeFormat = "2006-01-02 15:04:05 UTC"

const AdminUserId string = "admin"

type GetBalanceResult struct {
	Status  Status  `json:"status"`
	Balance Balance `json:"balance"`
}

type GetHistoryResult struct {
	Status  Status                         `json:"status"`
	History []TransferHistoryWithTimestamp `json:"history"`
}

type TransferResult struct {
	Status          Status  `json:"status"`
	FromUserBalance Balance `json:"from_user_balance"`
	ToUserBalance   Balance `json:"to_user_balance"`
}

type IssueNewPointResult struct {
	Status       Status  `json:"status"`
	AdminBalance Balance `json:"admin_balance"`
}

type ErrorResult struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// Define the Smart Contract structure
type SmartContract struct {
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return s.initAdmin(APIstub)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "getBalance" {
		return s.getBalance(APIstub, args)
	}
	if function == "getHistory" {
		return s.getHistory(APIstub, args)
	}
	if function == "transfer" {
		return s.transfer(APIstub, args)
	}
	if function == "issueNewPoint" {
		return s.issueNewPoint(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// Balance用Stateキー作成関数
func (s *SmartContract) makeBalanceKey(userId string) string {
	return "balance_" + userId
}

// TransferBill用Stateキー作成関数
func (s *SmartContract) makeTransferBillKey(userId string, month string) string {
	return "transfer_bill_" + userId + "_" + month
}

// 残高が空のBalance構造体作成
func (s *SmartContract) makeEmptyBalance(userId string) Balance {
	result := Balance{
		UserId: userId,
		Amount: 0.0,
		Total:  0.0,
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

// 指定ユーザーの残高取得
func (s *SmartContract) getUserBalance(APIstub shim.ChaincodeStubInterface, userId string) Balance {
	key := s.makeBalanceKey(userId)

	balanceAsBytes, _ := APIstub.GetState(key)
	balance := new(Balance)

	if len(balanceAsBytes) != 0 {
		json.Unmarshal(balanceAsBytes, balance)
	}

	return *balance
}

// 指定残高データが有効であるか判別する
func (s *SmartContract) isValidBalance(balance Balance) bool {
	if balance.UserId == "" {
		return false
	}

	// 論理削除や有効期限など追加した場合はここでチェックする

	return true
}

// Balanceデータput
func (s *SmartContract) putBalance(APIstub shim.ChaincodeStubInterface, balance Balance) {
	key := s.makeBalanceKey(balance.UserId)
	balanceAsBytes, _ := json.Marshal(balance)
	APIstub.PutState(key, balanceAsBytes)
}

// 取引明細データput
func (s *SmartContract) putTransferBill(APIstub shim.ChaincodeStubInterface, userId string, history TransferHistory) {
	month := time.Now().Format(MonthFormat)
	key := s.makeTransferBillKey(userId, month)
	historyAsBytes, _ := json.Marshal(history)
	APIstub.PutState(key, historyAsBytes)
}

// 管理者残高データの初期化
func (s *SmartContract) initAdmin(APIstub shim.ChaincodeStubInterface) sc.Response {
	adminBalance := s.getUserBalance(APIstub, AdminUserId)
	// すでに有効な残高データがあるならそれを返却
	if s.isValidBalance(adminBalance) {
		result := GetBalanceResult{Status: StatusOk, Balance: adminBalance}
		resultAsBytes, _ := json.Marshal(result)
		return shim.Success(resultAsBytes)
	}

	// 有効な残高がなければ空のデータを作成してput
	emptyBalance := s.makeEmptyBalance(AdminUserId)
	s.putBalance(APIstub, emptyBalance)

	result := GetBalanceResult{Status: StatusOk, Balance: emptyBalance}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := s.makeBalanceKey(args[0])
	balanceAsBytes, _ := APIstub.GetState(key)
	balance := Balance{}
	json.Unmarshal(balanceAsBytes, &balance)

	result := GetBalanceResult{Status: StatusOk, Balance: balance}
	if balance.UserId == "" {
		result.Status = StatusNotFound
	}

	resultAsBytes, _ := json.Marshal(result)
	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userId := args[0]
	yearMonth := args[1]
	key := s.makeTransferBillKey(userId, yearMonth)

	historyIterator, err := APIstub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyIterator.Close()

	var transferHistories []TransferHistoryWithTimestamp
	for historyIterator.HasNext() {
		queryResponse, err := historyIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		transferHistory := TransferHistory{}
		json.Unmarshal(queryResponse.Value, &transferHistory)
		transferHistoryWithTimestamp := TransferHistoryWithTimestamp{transferHistory, time.Unix(queryResponse.Timestamp.Seconds, 0).Format(DateTimeFormat)}
		transferHistories = append(transferHistories, transferHistoryWithTimestamp)
	}

	result := GetHistoryResult{Status: StatusOk, History: transferHistories}
	if len(transferHistories) < 1 {
		result.Status = StatusNotFound
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) transfer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	fromUserId := args[0]
	toUserId := args[1]
	point, err := strconv.ParseFloat(args[2], 32)
	if err != nil {
		return shim.Error("Incorrect type of arguments.")
	}

	// 譲渡元の残高取得
	fromBalance := s.getUserBalance(APIstub, fromUserId)

	// 有効チェック
	if !s.isValidBalance(fromBalance) {
		return s.makeErrorResponce(APIstub, StatusNotFound, "譲渡元の残高が見つかりませんでした")
	}

	// 手数料など追加するならここで
	total := float32(point)

	// 譲渡元の残高が足りているかチェック
	if total > fromBalance.Amount {
		return s.makeErrorResponce(APIstub, StatusBadRequest, "譲渡元の残高が足りません")
	}

	// 譲渡先の残高取得
	toBalance := s.getUserBalance(APIstub, toUserId)
	// なければ初期化
	if !s.isValidBalance(toBalance) {
		toBalance = s.makeEmptyBalance(toUserId)
	}

	// 譲渡先の残高増加
	toBalance.Amount += total
	toBalance.Total += total
	s.putBalance(APIstub, toBalance)

	// 譲渡先の当月最新取引データ更新
	fromUserHistory := TransferHistory{
		ToUserId:   toUserId,
		FromUserId: fromUserId,
		Point:      -total,
	}
	s.putTransferBill(APIstub, fromUserId, fromUserHistory)

	// 譲渡元の残高減算
	fromBalance.Amount -= total
	s.putBalance(APIstub, fromBalance)

	// 譲渡元の当月最新取引データ更新
	toUserHistory := TransferHistory{
		ToUserId:   toUserId,
		FromUserId: fromUserId,
		Point:      total,
	}
	s.putTransferBill(APIstub, toUserId, toUserHistory)

	// レスポンス作成
	result := TransferResult{
		Status:          StatusOk,
		FromUserBalance: fromBalance,
		ToUserBalance:   toBalance,
	}
	resultAsBytes, _ := json.Marshal(result)

	// 返却
	return shim.Success(resultAsBytes)
}

func (s *SmartContract) issueNewPoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	val, err := strconv.ParseFloat(args[0], 32)
	if err != nil {
		return shim.Error("Incorrect type of arguments.")
	}
	point := float32(val)

	// 管理者の残高の取得
	adminBalanceKey := s.makeBalanceKey(AdminUserId)
	balanceAsBytes, _ := APIstub.GetState(adminBalanceKey)
	adminBalance := Balance{}
	json.Unmarshal(balanceAsBytes, &adminBalance)

	// balanceを更新
	adminBalance.Amount += point
	adminBalance.Total += point
	s.putBalance(APIstub, adminBalance)

	// 履歴を更新
	userHistory := TransferHistory{
		ToUserId:   AdminUserId,
		FromUserId: AdminUserId,
		Point:      point,
	}
	s.putTransferBill(APIstub, AdminUserId, userHistory)

	result := IssueNewPointResult{Status: StatusOk, AdminBalance: adminBalance}
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
