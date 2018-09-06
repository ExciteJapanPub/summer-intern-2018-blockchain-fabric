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

const DateTimeFormat = "2006-01-02 15:04:05 UTC"
type Room struct {
  Id              string    `json:"id"`
  StatusOfUse     string    `json:"status_of_use"`
  UnreservingTime time.Time `json:"unreserving_time"`
}

type Status int
const (
	StatusOk Status = 200
	StatusCreated Status = 201
	StatusConflict Status = 409
)

type ResultUser struct {
  Status  Status  `json:"status"`
  User    User    `json:"user"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
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

	return shim.Error("Invalid Smart Contract function name.")
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

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
