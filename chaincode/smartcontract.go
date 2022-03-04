package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing user information
type SmartContract struct {
	contractapi.Contract
}

// User describes basic details of what makes up a user
type User struct {
	ID              string `json:"id"`
	IMSI            string `json:"imsi"`
	ICCID           string `json:"iccid"`
	ServiceNumber   string `json:"service_number"`
	ServicePassword string `json:"service_password"`
	PublicKey       string `json:"public_key"`
	Authority       string `json:"authority"`
}

// InitLedger adds a base set of users to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var (
		imsi            = "460001234567890"
		iccid           = "898600f92952f2340807"
		serviceNumber   = "15216730052"
		servicePassword = "123456"
	)

	//idBytes := sha256.Sum256([]byte(imsi))
	//id := hex.EncodeToString(hashedImsiByte[:])
	testID := "initLedgerTest-ID"

	users := []User{
		{
			ID:              testID,
			IMSI:            imsi,
			ICCID:           iccid,
			ServiceNumber:   serviceNumber,
			ServicePassword: servicePassword,
		},
	}

	for _, user := range users {
		userJSON, _ := json.Marshal(user)
		err := ctx.GetStub().PutState(user.ID, userJSON)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateUser adds a new user to the world state with given details
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, imsi string,
	iccid string, serviceNumber string, servicePassword string) error {
	//idBytes := sha256.Sum256([]byte(imsi))
	//id := hex.EncodeToString(idBytes[:])
	testID := "createUserTest-ID"

	user := User{
		ID:              testID,
		IMSI:            imsi,
		ICCID:           iccid,
		ServiceNumber:   serviceNumber,
		ServicePassword: servicePassword,
	}

	userJSON, _ := json.Marshal(user)

	return ctx.GetStub().PutState(testID, userJSON)
}

// DeleteUser deletes a given user from the world state
func (s *SmartContract) DeleteUser(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user with given id <%s> does not exists", id)
	}

	return ctx.GetStub().DelState(id)
}

// UserExists returns true when user with given ID exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

// GetUser returns the user stored in the world state with given id (h-imsi)
func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %v", err)
	}

	if userJSON == nil {
		return nil, fmt.Errorf("user does not exist")
	}

	user := new(User)
	_ = json.Unmarshal(userJSON, user)

	return user, nil
}

// GetAllUsers returns all users found in world state
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*User

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		user := new(User)
		err = json.Unmarshal(queryResponse.Value, user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// RegisterUser registers user by adding user's public_key and authority into the world state
func (s *SmartContract) RegisterUser(ctx contractapi.TransactionContextInterface, serviceNumber string,
	servicePassword string, publicKey string, authority string) error {

	users, err := s.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.ServiceNumber == serviceNumber && user.ServicePassword == servicePassword {
			user.PublicKey = publicKey
			user.Authority = authority
			userJSON, _ := json.Marshal(user)
			return ctx.GetStub().PutState(user.ID, userJSON)
		}
	}

	return fmt.Errorf("register failed, user not found. please check the service_number or the service_password provided")
}

// ChangeUserServicePassword updates the ServicePassword field of user with given id in world state
func (s *SmartContract) ChangeUserServicePassword(ctx contractapi.TransactionContextInterface,
	id string, newServicePassword string) error {
	user, err := s.GetUser(ctx, id)

	if err != nil {
		return err
	}

	user.ServicePassword = newServicePassword

	userJSON, _ := json.Marshal(user)

	return ctx.GetStub().PutState(id, userJSON)
}
