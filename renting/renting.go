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
	//"strconv"
	//"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)

//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4
const   AUTHORITY      =  "regulator"
const   HOUSE_OWNER	 =  "house_owner"
const   AGENT_COMPANY  =  "agent_company"
const   LEASEE		=  "leasee"

//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 4 statuses, this is part of the business logic to determine what can
//					be done to the house at points in it's lifecycle
//==============================================================================================================================
const   STATE_TEMPLATE  	=  0
const   STATE_HOUSEOWNER  	=  1
const   STATE_AGENT 		=  2
const   STATE_LEASEE 		=  3

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type House struct{
  Address	string	`json:"address"`
  Owner		string	`json:"owner"`
  Status 	int	`json:"status"`
  HouseID	string	`json:"houseID"`	
  Money		string	`json:"money"`
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

	caller := args[0]
	caller_affiliation := args[1]
        //caller, caller_affiliation, err := t.get_caller_data(stub)
        //if err != nil { return nil, errors.New("Error retrieving caller information")}

	// Handle different functions
	if function == "init" {	//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function =="create_house" {
		return t.create_house(stub, caller, caller_affiliation, args[2], args[3])
        } else if function == "invoke_agent_to_leasee_111" {
                h, err := t.retrieve_house(stub, args[2])
                if err != nil {
                        fmt.Printf("QUERY: Error retrieving house : %s", err)
                        return nil, errors.New("QUERY: Error retrieving house "+err.Error())
                }
                return t.agent_to_leasee(stub, h, args[0], AGENT_COMPANY, args[1], LEASEE, args[3] )
        } else { 
	        argPos := 2
        	h, err := t.retrieve_house(stub, args[argPos])
        	if err != nil {
                	fmt.Printf("INVOKE: Error retrieving house: %s", err)
                	return nil, errors.New("Error retrieving house")
        	}

		if function == "authority_to_houseowner" {
                	return t.authority_to_houseowner(stub, h, args[0], AUTHORITY, args[1], HOUSE_OWNER )
        	} else if  function == "houseowner_to_agent"   {
                	return t.houseowner_to_agent(stub, h, args[0], HOUSE_OWNER, args[1], AGENT_COMPANY, args[3] )
        	} else if  function == "agent_to_leasee" {
                	return t.agent_to_leasee(stub, h, args[0], AGENT_COMPANY, args[1], LEASEE, args[3] )
        	}
	}

	fmt.Println("invoke did not find func: " + function)	//error

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
func (t *SimpleChaincode) create_house(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, _houseID string, _address string) ([]byte, error) {
	var h House

	house_ID       := "\"HouseID\":\"" + _houseID + "\","	// Variables to define the JSON
	owner          := "\"Owner\":\"" + caller + "\","
	address	       := "\"Address\":\"" +_address + "\","
	status         := "\"Status\":0,"
	money         := "\"Money\":\"UNDEFINED\""

	house_json := "{"+house_ID+owner+address+status+money+"}" 	// Concatenates the variables to create the total JSON object
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(_houseID))  // matched = true if the _houseID passed fits format of two letters followed by seven digits
	if err != nil {
   		 fmt.Printf("CREATE_HOUSE: Invalid _houseID: %s", err); return nil, errors.New("Invalid _houseID")
  	}

	if _houseID == "" || matched == false {
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

	_, err  = t.save_changes(stub, h)
	if err != nil {
    		fmt.Printf("CREATE_HOUSE: Error saving changes: %s", err)
   		return nil, errors.New("Error saving changes")
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

	houseIDs.HIDs= append(houseIDs.HIDs, _houseID)
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
        } else if function == "ping" {
                return t.ping(stub)
	} else if function == "read" {
		return t.read(stub, args)
	} else if function == "get_house_details" {
		if len(args) != 3 { 
			fmt.Printf("Incorrect number of arguments passed") 
			return nil, errors.New("QUERY: Incorrect number of arguments passed") 
		}
		h, err := t.retrieve_house(stub, args[2])
		if err != nil { 
			fmt.Printf("QUERY: Error retrieving house : %s", err) 
			return nil, errors.New("QUERY: Error retrieving house "+err.Error()) 
		}
                return t.get_house_details(stub, args[0], args[1], h)
        } else if function == "get_houses" {
                return t.get_houses(stub, args[0], args[1])
	} else if function == "invoke_agent_to_leasee" {
                h, err := t.retrieve_house(stub, args[2])
                if err != nil {
                        fmt.Printf("QUERY: Error retrieving house : %s", err)
                        return nil, errors.New("QUERY: Error retrieving house "+err.Error())
                }
                return t.agent_to_leasee(stub, h, args[0], AGENT_COMPANY, args[1], LEASEE, args[3] )
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

//==============================================================================================================================
// save_changes - Writes to the ledger the House struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, h House) (bool, error) {
	bytes, err := json.Marshal(h)
	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error converting house record: %s", err) 
		return false, errors.New("Error converting house record") 
	}
	err = stub.PutState(h.HouseID, bytes)
	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error storing house record: %s", err) 
		return false, errors.New("Error storing house record") 
	}
	return true, nil
}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 authority_to_houseowner
//=================================================================================================================================
func (t *SimpleChaincode) authority_to_houseowner(stub shim.ChaincodeStubInterface, h House, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if  h.Status	== STATE_TEMPLATE	&&
		caller_affiliation	== AUTHORITY		&&
		recipient_affiliation	== HOUSE_OWNER {		// If the roles and users are ok

		h.Owner  = recipient_name		// then make the owner the new owner
		h.Status = STATE_HOUSEOWNER // and mark it in the state of house owner
	} else {	// Otherwise if there is an error
		fmt.Printf("authority_to_houseowner: Permission Denied")
                return nil, errors.New(fmt.Sprintf("Permission Denied. authority_to_houseowner"))
	}

	_, err := t.save_changes(stub, h)						// Write new state

	if err != nil {	
		fmt.Printf("authority_to_houseowner: Error saving changes: %s", err) 
		return nil, errors.New("Error saving changes")	
	}
	return nil, nil									// We are Done
}

//=================================================================================================================================
//       houseowner_to_agent
//=================================================================================================================================
func (t *SimpleChaincode) houseowner_to_agent(stub shim.ChaincodeStubInterface, h House, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string, _money string) ([]byte, error) {

        if  h.Status == STATE_HOUSEOWNER       &&
          	h.Owner	== caller			&&
                caller_affiliation == HOUSE_OWNER            &&
                recipient_affiliation == AGENT_COMPANY {                // If the roles and users are ok

                h.Owner  = recipient_name               // then make the owner the new owner
                h.Status = STATE_AGENT // and mark it in the state of house owner
		h.Money = _money 
        } else {        // Otherwise if there is an error
                fmt.Printf("houseowner_to_agent: Permission Denied")
                return nil, errors.New(fmt.Sprintf("Permission Denied. houseowner_to_agent"))
        }

        _, err := t.save_changes(stub, h)                                               // Write new state

        if err != nil {
                fmt.Printf("houseowner_to_agent: Error saving changes: %s", err)
                return nil, errors.New("Error saving changes")
        }
        return nil, nil                                                                 // We are Done
}

//=================================================================================================================================
//       agent_to_leasee
//=================================================================================================================================
func (t *SimpleChaincode) agent_to_leasee(stub shim.ChaincodeStubInterface, h House, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string, _money string) ([]byte, error) {

        if  h.Status == STATE_AGENT       &&
          	h.Owner	== caller			&&
                caller_affiliation == AGENT_COMPANY            &&
                recipient_affiliation == LEASEE {                // If the roles and users are ok

                h.Owner  = recipient_name               // then make the owner the new owner
                h.Status = STATE_LEASEE // and mark it in the state of house owner
		h.Money = _money
        } else {        // Otherwise if there is an error
                fmt.Printf("agent_to_leasee: Permission Denied")
                return nil, errors.New(fmt.Sprintf("Permission Denied. agent_to_leasee"))
        }

        _, err := t.save_changes(stub, h)                                               // Write new state

        if err != nil {
                fmt.Printf("agent_to_leasee: Error saving changes: %s", err)
                return nil, errors.New("Error saving changes")
        }
        return nil, nil                                                                 // We are Done
}

//==============================================================================================================================
//	 retrieve_house - Gets the state of the data at houseID in the ledger then converts it from the stored
//					JSON into the House struct for use in the contract. Returns the House struct.
//					Returns empty h if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_house(stub shim.ChaincodeStubInterface, houseID string) (House, error) {
	var h House
	bytes, err := stub.GetState(houseID);
	if err != nil {	
		fmt.Printf("RETRIEVE_HOUSE: Failed to invoke house_code: %s", err)
		return h , errors.New("RETRIEVE_HOUSE: Error retrieving house with houseID = " + houseID) 
	}
	err = json.Unmarshal(bytes, &h)
  	if err != nil {	
		fmt.Printf("RETRIEVE_HOUSE: Corrupt house record "+string(bytes)+": %s", err) 
 		 return h, errors.New("RETRIEVE_HOUSE: Corrupt house record"+string(bytes))	
	}
	return h, nil
}

//=================================================================================================================================
//	 get_houses
//=================================================================================================================================
func (t *SimpleChaincode) get_houses(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string) ([]byte, error) {
	bytes, err := stub.GetState("HouseIDs")
  	if err != nil { 
		return nil, errors.New("Unable to get HouseIDs") 
	}

	var houseIDs House_Holder
	err = json.Unmarshal(bytes, &houseIDs)
	if err != nil {	
		return nil, errors.New("Corrupt House_Holder") 
	}

	result := "["
	var temp []byte
	var h House
	for _, hid := range houseIDs.HIDs {
		h, err = t.retrieve_house(stub, hid)
    		if err != nil {
			return nil, errors.New("Failed to retrieve HID")
		}
		temp, err = t.get_house_details(stub, caller, caller_affiliation, h)
		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 get_house_details
//=================================================================================================================================
func (t *SimpleChaincode) get_house_details(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, h House) ([]byte, error) {
	bytes, err := json.Marshal(h)
	if err != nil { 
		return nil, errors.New("GET_HOUSE_DETAILS: Invalid house object") 
	}
	if h.Owner	== caller		||
		caller_affiliation == AUTHORITY	{
		return bytes, nil
	} else {
		return nil, errors.New("Permission Denied. get_house_details")
	}
}

//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}
