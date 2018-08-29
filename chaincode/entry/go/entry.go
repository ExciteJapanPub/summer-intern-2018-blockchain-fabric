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

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

type Entry struct {
	Value string `json:"value"` // "hoge"
	CreatedAt string `json:"created_at"`  // "2006-01-02 15:04:05"
}

type MonthlyEntries struct {
	Month string `json:"month"`                // "2006-01"
	DailyEntries  map[string]Entry `json:"daily_entries"`  // key: "2006-01-02"
}

type Status int
const (
	StatusOk Status = 200
	StatusCreated Status = 201
	StatusConflict Status = 409
)

type Result struct {
	Status      Status `json:"status"`
	Entries MonthlyEntries `json:"entries"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "putEntry" {
		return s.putEntry(APIstub, args)
	}
	if function == "getEntries" {
		return s.getEntries(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) getEntries(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ID
	no := args[0]
	// 対象月 "2006-01"
	month := args[1]

	// データ取得
	key := s.makeMonthlyEntriesKey(no, month)
	data := s.getMonthlyEntries(APIstub, key)

	// データがなければ返却用にmonthに指定月を入れた空のMonthlyEntriesを作成
	if data.Month == "" {
		data.Month = month
		data.DailyEntries = map[string]Entry{}
	}

	// 返却値生成
	result := Result{Status: StatusOk, Entries: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) putEntry(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ID
	no := args[0]
	// 出勤日時 "2006-01-02 15:04:05"
	entryTime := args[1]
	// 対象月 "2006-01"
	month := entryTime[0:7]
	// 対象日 "2006-01-02"
	day := entryTime[0:10]
	// Value
	value := args[2]

	// データ取得
	key := s.makeMonthlyEntriesKey(no, month)
	data := s.getMonthlyEntries(APIstub, key)

	// データ更新すべきかチェック
	if !s.shouldUpdate(data, day, value) {
		result := Result{Status: StatusConflict}
		resultAsBytes, _ := json.Marshal(result)
		return shim.Success(resultAsBytes)
	}

	// 指定月のデータが取得できていない場合は初期化
	if data.Month == "" {
		data.Month = month
		data.DailyEntries = map[string]Entry{}
	}

	// 指定日のデータ作成
	entry := Entry{Value: value, CreatedAt: entryTime}
	data.DailyEntries[day] = entry

	// データ登録
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(key, dataAsBytes)

	// 返却データ作成
	result := Result{Status: StatusCreated, Entries: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

// キー作成関数
func (s *SmartContract) makeMonthlyEntriesKey(no string, month string) string {
	return no + "_" + month
}

// データ取得用関数
func (s *SmartContract) getMonthlyEntries(APIstub shim.ChaincodeStubInterface, key string) MonthlyEntries {
	dataAsBytes, _ := APIstub.GetState(key)
	data := MonthlyEntries{}
	json.Unmarshal(dataAsBytes, &data)

	return data
}


// データを更新すべきかどうか
func (s *SmartContract) shouldUpdate(data MonthlyEntries, day string, value string) bool {
	// 指定日のデータが作成済みかチェック(作成されていなければ更新すべきなのでtrue)
	_, exists := data.DailyEntries[day]
	if !exists {
		return true;
	}

	// TODO 更新するかどうかの条件をここに記述
	// ひとまずvalueが空でなければ更新
	return (value != "")
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
