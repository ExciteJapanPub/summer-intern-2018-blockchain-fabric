package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)


type SmartContract struct {

}

type User struct {
  Id              string `json:"id"`
  Password        string `json:"password"`
  Balance         int    `json:"balance"`
  ReservedRoomId  string `json:"reserved_room_id"`
}

type Room struct {
  Id              string    `json:"id"`
  StatusOfUse     string    `json:"status_of_use"`
}
type Status int
const (
	StatusOk Status = 200
	StatusCreated Status = 201
  StatusNotFound Status = 404
	StatusConflict Status = 409
)

type ResultRoom struct {
  Status  Status  `json:"status"`
  Room    Room    `json:"room"`
}

type ResultAllRooms struct {
  Status  Status  `json:"status"`
  Rooms    []Room    `json:"rooms"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()

	if function == "putRoom" {
		return s.putRoom(APIstub, args)
	}
	if function == "getRoom" {
		return s.getRoom(APIstub, args)
	}
  if function == "getAllRooms" {
		return s.getAllRooms(APIstub)
  }
	if function == "putUser" {
			return s.putUser(APIstub, args)
	}
	if function == "getUser" {
		return s.getUser(APIstub, args)
	}
	if function == "updateReservedRoomId" {
		return s.updateReservedRoomId(APIstub, args)
	}
	if function == "updateBalance" {
		return s.updateBalance(APIstub, args)
}

func (s *SmartContract) getRoom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// データ取得
	key := args[0]

	// 返却値生成
  dataAsBytes, _ := APIstub.GetState(key)
  data := Room{}
  json.Unmarshal(dataAsBytes, &data)

	result := ResultRoom{Status: StatusOk, Room: data}

  return shim.Success(resultAsBytes)
}

func (s *SmartContract) getUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("")
	}

	password := args[0]

	key := password
	dataAsBytes, _ := APIstub.GetState(key)
	data := User{}
	json.Unmarshal(dataAsBytes, &data)

	result := ResultUser{Status: StatusOk, User: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getAllRooms(APIstub shim.ChaincodeStubInterface) sc.Response {
  startKey := "Room0"
  endKey := "Room999"

  resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
  if err != nil {
    return shim.Error(err.Error())
  }
  defer resultsIterator.Close()

  var rooms []Room
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		room := Room{}
		json.Unmarshal(queryResponse.Value, &room)
		rooms = append(rooms, room)
	}

  // 返却値生成
	result := ResultAllRooms{Status: StatusOk, Rooms: rooms}
	if len(rooms) < 1 {
		result.Status = StatusNotFound
	}

  return shim.Success(resultAsBytes)
}

func (s *SmartContract) putUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("")
	}

	no := args[0]
	password := args[1]
	balance := 0
	reservedRoomId := ""

	key := password
	data := User{no, password, balance, reservedRoomId}

	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(key, dataAsBytes)

	result := ResultUser{Status: StatusCreated, User: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) putRoom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	RoomId := args[0]
	StatusOfUse := "notUsed"

	key := RoomId
	data := Room{Id: RoomId, StatusOfUse: StatusOfUse}

	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(key, dataAsBytes)

	result := ResultRoom{Status: StatusCreated, Room: data}
	resultAsBytes, _ := json.Marshal(result)

  return shim.Success(resultAsBytes)
}

func (s *SmartContract) updateReservedRoomId(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("")
	}

	password := args[0]
	reservedRoomId := args[1]

	key := password
	dataAsBytes, _ := APIstub.GetState(key)
	data := User{}
	json.Unmarshal(dataAsBytes, &data)

	if data.Id != "" {
		data.ReservedRoomId = reservedRoomId

		dataAsBytes, _ := json.Marshal(data)
		APIstub.PutState(key, dataAsBytes)
	}

	result := ResultUser{Status: StatusOk, User: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) updateBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("")
	}

	password := args[0]
	balance, _ := strconv.Atoi(args[1])

	key := password
	dataAsBytes, _ := APIstub.GetState(key)
	data := User{}
	json.Unmarshal(dataAsBytes, &data)

	if data.Id != "" {
		data.Balance = balance

		dataAsBytes, _ := json.Marshal(data)
		APIstub.PutState(key, dataAsBytes)
	}

	result := ResultUser{Status: StatusOk, User: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) getUserFromStateDB(APIstub shim.ChaincodeStubInterface, password string) User {
	key := password
	dataAsBytes, _ := APIstub.GetState(key)
	data := User{}
	json.Unmarshal(dataAsBytes, &data)

	return data
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
