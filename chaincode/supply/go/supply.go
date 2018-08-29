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

type ResultStatus int

const (
	StatusOk         ResultStatus = 200
	StatusCreated    ResultStatus = 201
	StatusBadRequest ResultStatus = 400
	StatusNotFound   ResultStatus = 404
)

type DeliveryStatus string

const (
	StatusReceivedOrder DeliveryStatus = "ordered"
	StatusOnPassage     DeliveryStatus = "on_passage"
	StatusDelivered     DeliveryStatus = "delivered"
)

const StripDateTimeFormat = "20060102150405"

type Item struct {
	ItemId string `json:"item_id"` // "item_xxx"
	Name   string `json:"name"`    // "apple"
	Stock  int    `json:"stock"`   // 100
}

type Delivery struct {
	DeliveryId string         `json:"delivery_id"` // "delivery_userId_itemId_YmdHis"
	UserId     string         `json:"user_id"`     // "1234"
	ItemId     string         `json:"item_id"`     // "item_xxx"
	Quantity   int            `json:"quatity"`     // 100
	Status     DeliveryStatus `json:"status"`      // "ordered"
}

// TODO 実際にサービス化するなら未配送の商品と月ごとの発送済み商品で分けるなどして、DeliveryIdsが大きくなりすぎないようにする
type UserDeliveries struct {
	UserId      string   `json:"user_id"` // "1234"
	DeliveryIds []string `json:"delivery_ids"`
}

type ItemResult struct {
	Status ResultStatus `json:"status"`
	Item   Item         `json:"item"`
}

type DeliveryResult struct {
	Status   ResultStatus `json:"status"`
	Delivery Delivery     `json:"delivery"`
}

type UserDeliveriesResult struct {
	Status     ResultStatus `json:"status"`
	UserId     string       `json:"user_id"` // "1234"
	Deliveries []Delivery   `json:"deliveries"`
}

type ErrorResult struct {
	Status  ResultStatus `json:"status"`
	Message string       `json:"message"`
}

// Define the Smart Contract structure
type SmartContract struct {
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	// 商品作成
	if function == "putItem" {
		return s.putItem(APIstub, args)
	}
	// 商品情報取得
	if function == "getItem" {
		return s.getItem(APIstub, args)
	}

	// 商品在庫補充
	if function == "replenishItem" {
		return s.replenishItem(APIstub, args)
	}

	// 商品購入
	if function == "buy" {
		return s.buy(APIstub, args)
	}

	// 配送状況の更新
	if function == "updateDeliveryStatus" {
		return s.updateDeliveryStatus(APIstub, args)
	}

	// ユーザーの配送データ取得
	if function == "getUserAllDeliveries" {
		return s.getUserAllDeliveries(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// Delivery.DeliveryIdの生成
func (s *SmartContract) generateDeliveryId(userId string, itemId string) string {
	datetime := time.Now().Format(StripDateTimeFormat)
	format := "delivery_%s_%s_%s"
	id := fmt.Sprintf(format, userId, itemId, datetime)

	return id
}

// Item保存用のキー作成
func (s *SmartContract) makeItemKey(itemId string) string {
	return "item_" + itemId
}

// UserDeliveries保存用のキー作成
func (s *SmartContract) makeUserDeliveriesKey(userId string) string {
	return "user_deliveries_" + userId
}

// Itemをstate DBから取得
func (s *SmartContract) getItemState(APIstub shim.ChaincodeStubInterface, itemId string) Item {
	key := s.makeItemKey(itemId)

	ItemAsBytes, _ := APIstub.GetState(key)
	item := new(Item)

	if len(ItemAsBytes) != 0 {
		json.Unmarshal(ItemAsBytes, item)
	}

	return *item
}

// Deliveryをstate DBから取得
func (s *SmartContract) getDelivery(APIstub shim.ChaincodeStubInterface, deliveryId string) Delivery {
	// keyはDeliveryIdをそのまま利用
	deliveryAsBytes, _ := APIstub.GetState(deliveryId)
	delivery := new(Delivery)

	if len(deliveryAsBytes) != 0 {
		json.Unmarshal(deliveryAsBytes, delivery)
	}

	return *delivery
}

// UserDeliveriesをstate DBから取得
func (s *SmartContract) getUserDeliveries(APIstub shim.ChaincodeStubInterface, userId string) UserDeliveries {
	key := s.makeUserDeliveriesKey(userId)

	userDeliveriesAsBytes, _ := APIstub.GetState(key)
	userDeliveries := new(UserDeliveries)

	if len(userDeliveriesAsBytes) != 0 {
		json.Unmarshal(userDeliveriesAsBytes, userDeliveries)
	}

	return *userDeliveries
}

// Itemをブロックチェーンにput
func (s *SmartContract) putItemToState(APIstub shim.ChaincodeStubInterface, item Item) {
	key := s.makeItemKey(item.ItemId)
	itemAsBytes, _ := json.Marshal(item)
	APIstub.PutState(key, itemAsBytes)
}

// Deliveryをブロックチェーンにput
func (s *SmartContract) putDelivery(APIstub shim.ChaincodeStubInterface, delivery Delivery) {
	// keyはDeliveryIdをそのまま利用
	key := delivery.DeliveryId
	deliveryAsBytes, _ := json.Marshal(delivery)
	APIstub.PutState(key, deliveryAsBytes)
}

// UserDeliveriesをブロックチェーンにput
func (s *SmartContract) putUserDeliveries(APIstub shim.ChaincodeStubInterface, deliveries UserDeliveries) {
	key := s.makeUserDeliveriesKey(deliveries.UserId)
	deliveriesAsBytes, _ := json.Marshal(deliveries)
	APIstub.PutState(key, deliveriesAsBytes)
}

// 変更可能な配送ステータスか判定
func (s *SmartContract) canChangeDeliveryStatus(fromStatus DeliveryStatus, toStatus DeliveryStatus) bool {
	var availableStatuses []DeliveryStatus

	switch fromStatus {
	case StatusReceivedOrder:
		availableStatuses = []DeliveryStatus{StatusOnPassage, StatusDelivered}
	case StatusOnPassage:
		availableStatuses = []DeliveryStatus{StatusDelivered}
	default:
		return false
	}

	for _, v := range availableStatuses {
		if v == toStatus {
			return true
		}
	}

	return false
}

// Deliveryデータの作成
func (s *SmartContract) makeDelivery(userId string, itemId string, quantity int) Delivery {
	deliveryId := s.generateDeliveryId(userId, itemId)
	delivery := Delivery{
		DeliveryId: deliveryId,
		UserId:     userId,
		ItemId:     itemId,
		Quantity:   quantity,
		Status:     StatusReceivedOrder,
	}

	return delivery
}

// 初期化したUserDeliveriesデータを作成
func (s *SmartContract) makeInitialUserDeliveries(userId string) UserDeliveries {
	ret := UserDeliveries{
		UserId:      userId,
		DeliveryIds: []string{},
	}

	return ret
}

// エラーレスポンスを作成
func (s *SmartContract) makeErrorResponce(APIstub shim.ChaincodeStubInterface, code ResultStatus, message string) sc.Response {
	result := ErrorResult{
		Status:  code,
		Message: message,
	}

	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) putItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	quantity, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Incorrect argument. third argument is should be numerical string")
	}

	item := Item{
		ItemId: args[0],
		Name:   args[1],
		Stock:  quantity,
	}
	s.putItemToState(APIstub, item)

	result := ItemResult{
		Status: StatusCreated,
		Item:   item,
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	item := s.getItemState(APIstub, args[0])
	if item.ItemId == "" {
		return s.makeErrorResponce(APIstub, StatusNotFound, "該当する商品が見つかりませんでした")
	}

	result := ItemResult{
		Status: StatusOk,
		Item:   item,
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) replenishItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	itemId := args[0]
	quantity, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Incorrect argument. second argument is should be numerical string")
	}

	item := s.getItemState(APIstub, itemId)
	if item.ItemId == "" {
		return s.makeErrorResponce(APIstub, StatusNotFound, "該当する商品が見つかりませんでした")
	}

	// アイテムを指定数分補充
	item.Stock += quantity
	// 保存
	s.putItemToState(APIstub, item)

	// 結果値作成
	result := ItemResult{
		Status: StatusOk,
		Item:   item,
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) buy(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	userId := args[0]
	itemId := args[1]
	quantity, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Incorrect argument. third argument is should be numerical string")
	}

	item := s.getItemState(APIstub, itemId)
	if item.ItemId == "" {
		return s.makeErrorResponce(APIstub, StatusNotFound, "該当する商品が見つかりませんでした")
	}
	if quantity > item.Stock {
		return s.makeErrorResponce(APIstub, StatusBadRequest, "在庫が足りません")
	}
	// 配送データ作成&put
	delivery := s.makeDelivery(userId, itemId, quantity)
	s.putDelivery(APIstub, delivery)

	// itemの在庫を減らす
	item.Stock -= quantity
	s.putItemToState(APIstub, item)

	// ユーザー配送データ取得
	userDelivery := s.getUserDeliveries(APIstub, userId)
	// 未作成なら初期化
	if userDelivery.UserId == "" {
		userDelivery = s.makeInitialUserDeliveries(userId)
	}

	// ユーザー配送データの配送IDリストに作成した配送データのIDを追加&put
	userDelivery.DeliveryIds = append(userDelivery.DeliveryIds, delivery.DeliveryId)
	s.putUserDeliveries(APIstub, userDelivery)

	// 結果値作成
	result := DeliveryResult{
		Status:   StatusCreated,
		Delivery: delivery,
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) updateDeliveryStatus(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	deliveryId := args[0]
	status := DeliveryStatus(args[1])

	delivery := s.getDelivery(APIstub, deliveryId)
	if delivery.DeliveryId == "" {
		return s.makeErrorResponce(APIstub, StatusNotFound, "該当する配送データが見つかりませんでした")
	}

	if !s.canChangeDeliveryStatus(delivery.Status, status) {
		return s.makeErrorResponce(APIstub, StatusBadRequest, "指定されたステータスには変更できません")
	}

	delivery.Status = status
	s.putDelivery(APIstub, delivery)

	// 結果値作成
	result := DeliveryResult{
		Status:   StatusOk,
		Delivery: delivery,
	}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getUserAllDeliveries(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	userId := args[0]

	userDeliveries := s.getUserDeliveries(APIstub, userId)
	if userDeliveries.UserId == "" {
		return s.makeErrorResponce(APIstub, StatusNotFound, "該当するユーザーの配送データが見つかりませんでした")
	}

	deliveries := []Delivery{}
	for _, id := range userDeliveries.DeliveryIds {
		delivery := s.getDelivery(APIstub, id)
		deliveries = append(deliveries, delivery)
	}

	// 結果値作成
	result := UserDeliveriesResult{
		Status:     StatusOk,
		UserId:     userId,
		Deliveries: deliveries,
	}
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
