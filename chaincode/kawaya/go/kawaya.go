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
	StatusConflict Status = 409
)

// type ResultUser struct {
//   Status  Status  `json:"status"`
//   User    User    `json:"user"`
// }

// const DateTimeFormat = "2006-01-02 15:04:05 UTC"
type Room struct {
  Id              string    `json:"id"`
  StatusOfUse     bool    `json:"status_of_use"`
  // UnreservingTime time.Time `json:"unreserving_time"`
}

type ResultRoom struct {
  Status  Status  `json:"status"`
  Room    Room    `json:"room"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	if function == "putRoom" {
			return s.putUser(APIstub, args)
	}
	if function == "getRoom" {
		return s.getUser(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) getRoom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// データ取得
	key := args[0]
	StatusOfUse := Room.StatusOfUse

	// 返却値生成
  dataAsBytes, _ := APIstub.GetState(key)
  data := Room{}
  json.Unmarshal(dataAsBytes, &data)

	result := ResultRoom{Status: StatusOk, Room: data}
	resultAsBytes, _ := json.Marshal(result)

	return shim.Success(resultAsBytes)
}

func (s *SmartContract) putRoom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	RoomId := args[0]
	StatusOfUse := false

	key := RoomId
	data := Room{Id, StatusOfUse}

	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(key, dataAsBytes)

	result := ResultUser{Status: StatusCreated, Room: data}
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
