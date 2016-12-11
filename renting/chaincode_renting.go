/*
Copyright VMware Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)

//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   AUTHORITY      =  "regulator"
const   MANUFACTURER   =  "manufacturer"
const   PRIVATE_ENTITY =  "private"
const   LEASE_COMPANY  =  "lease_company"
const   SCRAP_MERCHANT =  "scrap_merchant"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type House struct{
  Address	string	`json:"address"`
  Owner		string	`json:"owner"`
  Status 	int	`json:"status"`
  HouseID	string	`json:"houseID"`	
}

type House_Holder struct{
  HIDs	[]string `json:"hids"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
//	if len(args) != 1 {
//		return nil, errors.New("Incorrect number of arguments. Expecting 1")
//	}
	
	var houseIDs House_Holder
	bytes, err := json.Marshal(houseIDs)
	if err != nil {
		return nil, errors.New("Error creating House_Holder record")
	}

	err = stub.PutState("HouseIDs", bytes)


//	for i:=0; i < len(args); i=i+2 {
//		t.add_ecert(stub, args[i], args[i+1])
//	}
	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }

	return ecert, nil
}

//==============================================================================================================================
//	 add_ecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) add_ecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {


	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	var caller, caller_affiliation string

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function =="create_house" {
		rturn t.create_house(stub, caller, caller_affiliation, args[0], args[1])
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]
	value = args[1]
	err = stub.PutState(key, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create House - Creates the initial JSON for the house and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_vehicle(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, _houseID string, _address string) ([]byte, error) {
	var h House

	house_ID       := "\"HouseID\":\""+_houseID+"\", "							// Variables to define the JSON
	owner          := "\"Owner\":\""+caller+"\", "
	address	       := "\"Address\":\""+_address\", "
	status         := "\"Status\":0"

	house_json := "{"+house_ID+owner+address+status+"}" 	// Concatenates the variables to create the total JSON object
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(_houseID))  // matched = true if the _houseID passed fits format of two letters followed by seven digits
	if err != nil {
   		 fmt.Printf("CREATE_HOUSE: Invalid _houseID: %s", err); return nil, errors.New("Invalid _houseID")
  	}

	if v5c_ID == "" || matched == false {
		fmt.Printf("CREATE_HOUSE: Invalid houseID provided")
		return nil, errors.New("Invalid houseID provided")
	}

	err = json.Unmarshal([]byte(house_json), &h)	// Convert the JSON defined above into a house object for go
	if err != nil {
   		return nil, errors.New("Invalid JSON object")
 	}
	record, err := stub.GetState(h.HouseID)  // If not an error then a record exists so cant create a new house with this HouseID as it must be unique

	if record != nil {
    		return nil, errors.New("House already exists")
  	}

	if caller_affiliation != AUTHORITY {	 // Only the regulator can create a new house 
		return nil, errors.New(fmt.Sprintf("Permission Denied. create_house. %v === %v", caller_affiliation, AUTHORITY))
	}

	_, err  = t.save_changes(stub, v)
	if err != nil {
    		fmt.Printf("CREATE_HOUSE: Error saving changes: %s", err)
   		return nil, Errors.New("Error saving changes")
  	}

	bytes, err := stub.GetState("HouseIDs")
	if err != nil {
    		return nil, errors.New("Unable to get HouseIDs")
  	}

	var houseIDs House_Holder
	err = json.Unmarshal(bytes, &houseIDs)
	if err != nil {
    		return nil, errors.New("Corrupt House_Holder record")
  	}

	houseIDs.HIDs= append(houseIDs.HIDs, houseID)
	bytes, err = json.Marshal(houseIDs)
	if err != nil {
    		fmt.Print("Error creating House_Holder record")
  	}

	err = stub.PutState("HouseIDs", bytes)
	if err != nil {
    		return nil, errors.New("Unable to put the state")
  	}
	return nil, nil
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "dummy_query" {											//read a variable
		fmt.Println("hi there " + function)						//error
		return nil, nil;
	} else if function == "read" {
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error
	
	if len(args) != 1{
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	return valAsbytes, nil
}
