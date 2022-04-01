package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
	"time"
)

type SmartContract struct {
	contractapi.Contract
}

type tci contractapi.TransactionContextInterface

type Node struct {
	Id            string      `json:"id"`
	NodeType      string      `json:"nodeType"`
	PublicKeys    interface{} `json:"publicKeys"`
	AccessRecords interface{} `json:"accessRecords"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

// UserPublicKeys is the struct of user's public key
type UserPublicKeys map[string]string

// NodePair indicates the communication parties
type NodePair struct {
	MacAddr     string `json:"macAddr"`
	SatelliteId string `json:"satelliteId"`
}

// UserAccessRecord is single access log
type UserAccessRecord struct {
	AccessType          string    `json:"accessType"`          // 接入方式，normal/fast/handover
	PreviousSatelliteId string    `json:"previousSatelliteId"` // handover接入方式下，原卫星id
	StartAt             time.Time `json:"startAt"`             // 访问开始时间
	EndAt               time.Time `json:"endAt"`               // 访问结束时间
}

// UserAccessRecords indicates access records for a specific device
type UserAccessRecords map[NodePair][]UserAccessRecord

/**
********************************** smart contract implement ***************************************
 */

// InitLedger initialize the ledger
func (s *SmartContract) InitLedger(ctx tci) error {
	userPublicKeys := UserPublicKeys{
		"macAddr1": "publicKey1",
		"macAddr2": "publicKey2",
	}
	userAccessRecords := UserAccessRecords{
		{
			MacAddr:     "macAddr1",
			SatelliteId: "star-1",
		}: {
			{
				AccessType:          "normal",
				PreviousSatelliteId: "",
				StartAt:             time.Now(),
				EndAt:               time.Now(),
			},
			{
				AccessType:          "fast",
				PreviousSatelliteId: "",
				StartAt:             time.Now(),
				EndAt:               time.Now(),
			},
			{
				AccessType:          "handover",
				PreviousSatelliteId: "star-0",
				StartAt:             time.Now(),
				EndAt:               time.Now(),
			},
		},
	}
	nodes := []Node{
		{
			Id:         "star-1",
			NodeType:   "star",
			PublicKeys: "star-1-publicKey",
		},
		{
			Id:         "star-2",
			NodeType:   "star",
			PublicKeys: "star-2-publicKey",
		},
		{
			Id:            "user-1",
			NodeType:      "user",
			PublicKeys:    userPublicKeys,
			AccessRecords: userAccessRecords,
		},
	}

	for _, node := range nodes {
		nodeJSON, _ := json.Marshal(node)
		err := ctx.GetStub().PutState(node.Id, nodeJSON)

		if err != nil {
			return errors.Wrap(err, "failed to put into world state")
		}
	}

	return nil
}

func (s *SmartContract) SatelliteRegister(ctx tci, id string, publicKey string) error {
	_, err := s.GetNodeById(ctx, id)
	if err != nil {
		return err
	}

	satellite := &Node{
		Id:            id,
		NodeType:      "star",
		PublicKeys:    publicKey,
		AccessRecords: nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	satelliteJSON, _ := json.Marshal(satellite)

	return ctx.GetStub().PutState(id, satelliteJSON)
}

func (s *SmartContract) UserRegister(ctx tci, id string, macAddr string, publicKey string) error {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return err
	}

	var userJSON []byte

	if node.Id == "" {
		newUser := &Node{
			Id:            id,
			NodeType:      "user",
			PublicKeys:    map[string]string{macAddr: publicKey},
			AccessRecords: UserAccessRecords{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		userJSON, _ = json.Marshal(newUser)
	}

	if node.NodeType == "user" {
		if _, ok := node.PublicKeys.(UserPublicKeys)[macAddr]; ok {
			return errors.New("you've already registered")
		}
		node.PublicKeys.(UserPublicKeys)[macAddr] = publicKey
		node.UpdatedAt = time.Now()
		userJSON, _ = json.Marshal(node)

	}

	return ctx.GetStub().PutState(id, userJSON)
}

func (s *SmartContract) CreateAccessRecord(ctx tci, id string, macAddr string, satelliteId string, userAccessRecord UserAccessRecord) error {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return err
	}

	if node.NodeType != "user" {
		return errors.New("cannot add access record into a non-user type object")
	}
	if _, ok := node.PublicKeys.(UserPublicKeys)[macAddr]; !ok {
		return errors.Errorf("user with id %s and macAddr %s does not exist. please register first", id, macAddr)
	}
	nodePair := NodePair{macAddr, satelliteId}

	node.AccessRecords = append(node.AccessRecords.(UserAccessRecords)[nodePair], userAccessRecord)
	node.UpdatedAt = time.Now()

	nodeJSON, _ := json.Marshal(&node)

	return ctx.GetStub().PutState(id, nodeJSON)
}

func (s *SmartContract) GetSatellitePublicKey(ctx tci, id string) (string, error) {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return "", err
	}

	if node.NodeType != "star" {
		return "", errors.Errorf("cannot get satellite's public key with non-star id %s", id)
	}

	publicKey := node.PublicKeys.(string)

	return publicKey, nil
}

func (s *SmartContract) GetUserPublicKey(ctx tci, id string, macAddr string) (string, error) {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return "", err
	}

	if node.NodeType != "user" {
		return "", errors.Errorf("cannot get user's public key with non-user id %s", id)
	}

	publicKey, ok := node.PublicKeys.(UserPublicKeys)[macAddr]
	if !ok {
		return "", errors.Errorf("public key of the user with id %s and macAddr %s does not exist. please register first", id, macAddr)
	}

	return publicKey, nil
}

func (s *SmartContract) GetNodeById(ctx tci, id string) (Node, error) {
	nodeJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return Node{}, errors.Wrap(err, "failed to read from world state")
	}
	if nodeJSON == nil {
		return Node{}, errors.New("user does not exist")
	}

	node := Node{}
	_ = json.Unmarshal(nodeJSON, &node)

	return node, nil
}

func (s *SmartContract) GetAllNodes(ctx tci) ([]Node, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var nodes []Node

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		node := Node{}
		err = json.Unmarshal(queryResponse.Value, &node)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}
