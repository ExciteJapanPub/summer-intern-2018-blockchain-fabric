package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)


type SmartContract struct {
}

// type User struct {
//   Id              string `json:"id"`
//   Password        string `json:"password"`
//   Balance         int    `json:"balance"`
//   ReservedRoomId  string `json:"reserved_room_id"`
// }

type Status int
const (
	StatusOk Status = 200
	StatusCreated Status = 201
  StatusNotFound Status = 404
	StatusConflict Status = 409
)

// type ResultUser struct {
//   Status  Status  `json:"status"`
//   User    User    `json:"user"`
// }

// const DateTimeFormat = "2006-01-02 15:04:05 UTC"
type Room struct {
  Id              string    `json:"id"`
  StatusOfUse     string    `json:"status_of_use"`
  // UnreservingTime time.Time `json:"unreserving_time"`
}

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

	return shim.Error("Invalid Smart Contract function name.")
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

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
