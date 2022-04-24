package chaincode

import (
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

type SmartContract struct {
	contractapi.Contract
}

type tci contractapi.TransactionContextInterface

// PublicKeys indicates the structure how public keys are saved
type PublicKeys map[string]string

// UserAccessRecord is single access log
type UserAccessRecord struct {
	AccessType          string `json:"accessType"`          // "normal" || "fast" || "handover"
	SatelliteId         string `json:"satelliteId"`         // current satellite
	PreviousSatelliteId string `json:"previousSatelliteId"` // previous satellite in handover
	StartAt             string `json:"startAt"`             // when to start
	EndAt               string `json:"endAt"`               // when to end
}

// UserAccessRecords indicates access records for a specific device
type UserAccessRecords map[string][]UserAccessRecord

type Node struct {
	Id           string            `json:"id"`
	NodeType     string            `json:"nodeType"`
	PublicKey    PublicKeys        `json:"publicKey"`
	AccessRecord UserAccessRecords `json:"accessRecord"`
	Status       string            `json:"status"` // 认证通过与否 ok
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

const (
	TimeTemplate          = "2006-01-02 15:04:05"
	SatelliteMacAddrAlias = "elden"

	UserStatusCertified   = "certified"
	UserStatusUnCertified = "uncertified"
)

/**
********************************** smart contract implement ***************************************
 */

//InitLedger initialize the ledger
func (s *SmartContract) InitLedger(ctx tci) error {
	userPublicKeys := PublicKeys{
		"macAddr1": "publicKey1",
		"macAddr2": "publicKey2",
	}
	userAccessRecords := UserAccessRecords{
		"macAddr1": {
			{
				AccessType:          "normal",
				SatelliteId:         "satellite-1",
				PreviousSatelliteId: "",
				StartAt:             time.Now().Format(TimeTemplate),
				EndAt:               time.Now().Format(TimeTemplate),
			},
			{
				AccessType:          "fast",
				SatelliteId:         "satellite-1",
				PreviousSatelliteId: "",
				StartAt:             time.Now().Format(TimeTemplate),
				EndAt:               time.Now().Format(TimeTemplate),
			},
			{
				AccessType:          "handover",
				SatelliteId:         "satellite-1",
				PreviousSatelliteId: "satellite-0",
				StartAt:             time.Now().Format(TimeTemplate),
				EndAt:               time.Now().Format(TimeTemplate),
			},
		},
	}
	nodes := []Node{
		{
			Id:           "user-1",
			NodeType:     "user",
			PublicKey:    userPublicKeys,
			AccessRecord: userAccessRecords,
			Status:       UserStatusUnCertified,
			CreatedAt:    time.Now().Format(TimeTemplate),
			UpdatedAt:    time.Now().Format(TimeTemplate),
		},
		{
			Id:           "satellite-1",
			NodeType:     "satellite",
			PublicKey:    PublicKeys{SatelliteMacAddrAlias: "satellite-1-publicKey"},
			AccessRecord: UserAccessRecords{},
			CreatedAt:    time.Now().Format(TimeTemplate),
			UpdatedAt:    time.Now().Format(TimeTemplate),
		},
		{
			Id:           "satellite-2",
			NodeType:     "satellite",
			PublicKey:    PublicKeys{SatelliteMacAddrAlias: "satellite-2-publicKey"},
			AccessRecord: UserAccessRecords{},
			CreatedAt:    time.Now().Format(TimeTemplate),
			UpdatedAt:    time.Now().Format(TimeTemplate),
		},
	}

	for _, node := range nodes {
		nodeJSON, err := json.Marshal(node)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}

		err = ctx.GetStub().PutState(node.Id, nodeJSON)
		if err != nil {
			return errors.Wrap(err, "failed to put into world state")
		}
	}

	return nil
}

func (s *SmartContract) SatelliteRegister(ctx tci, id string, publicKey string) error {
	exists, err := s.IsNodeExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return errors.Errorf("the node %s already exists", id)
	}

	satellite := Node{
		Id:           id,
		NodeType:     "satellite",
		PublicKey:    PublicKeys{SatelliteMacAddrAlias: publicKey},
		AccessRecord: UserAccessRecords{},
		CreatedAt:    time.Now().Format(TimeTemplate),
		UpdatedAt:    time.Now().Format(TimeTemplate),
	}

	satelliteJSON, err := json.Marshal(satellite)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	return ctx.GetStub().PutState(id, satelliteJSON)
}

func (s *SmartContract) UserRegister(ctx tci, id string, macAddr string, publicKey string) error {
	exists, err := s.IsNodeExists(ctx, id)
	if err != nil {
		return err
	}

	var userJSON []byte

	if exists {
		node, err := s.GetNodeById(ctx, id)
		if err != nil {
			return err
		}
		if node.NodeType != "user" {
			return errors.New("failed to call 'UserRegister' with provided non-user type")
		}
		if _, ok := node.PublicKey[macAddr]; ok {
			return errors.New("you've already registered (public key exists)")
		}
		node.PublicKey[macAddr] = publicKey
		node.UpdatedAt = time.Now().Format(TimeTemplate)

		userJSON, err = json.Marshal(*node)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}
	} else {
		newUser := Node{
			Id:           id,
			NodeType:     "user",
			PublicKey:    map[string]string{macAddr: publicKey},
			AccessRecord: UserAccessRecords{},
			Status:       UserStatusUnCertified,
			CreatedAt:    time.Now().Format(TimeTemplate),
			UpdatedAt:    time.Now().Format(TimeTemplate),
		}
		userJSON, err = json.Marshal(newUser)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}
	}

	return ctx.GetStub().PutState(id, userJSON)
}

func (s *SmartContract) CreateAccessRecord(ctx tci, id string, macAddr string, userAccessRecordString string) error {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return err
	}

	if node.NodeType != "user" {
		return errors.New("cannot add access record into a non-user type object")
	}
	if _, ok := node.PublicKey[macAddr]; !ok {
		return errors.Errorf("user with id %s and macAddr %s does not exist. please register first", id, macAddr)
	}

	var userAccessRecord UserAccessRecord
	err = json.Unmarshal([]byte(userAccessRecordString), &userAccessRecord)
	if err != nil {
		return errors.Wrap(err, "json unmarshal error")
	}

	node.AccessRecord[macAddr] = append(node.AccessRecord[macAddr], userAccessRecord)
	node.UpdatedAt = time.Now().Format(TimeTemplate)

	nodeJSON, err := json.Marshal(*node)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	return ctx.GetStub().PutState(id, nodeJSON)
}

func (s *SmartContract) ChangeAuthStatus(ctx tci, id string) error {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return err
	}

	if node.NodeType != "user" {
		return errors.New("cannot change auth-status for a non-user type object")
	}

	node.Status = UserStatusCertified
	node.UpdatedAt = time.Now().Format(TimeTemplate)

	nodeJSON, err := json.Marshal(*node)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	return ctx.GetStub().PutState(id, nodeJSON)
}

func (s *SmartContract) DeleteNodeById(ctx tci, id string) error {
	exists, err := s.IsNodeExists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("the node does not exist")
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) GetSatellitePublicKey(ctx tci, id string) (string, error) {
	node, err := s.GetNodeById(ctx, id)
	if err != nil {
		return "", err
	}

	if node.NodeType != "satellite" {
		return "", errors.Errorf("cannot get satellite's public key with non-satellite id %s", id)
	}

	publicKey := node.PublicKey[SatelliteMacAddrAlias]

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

	publicKey, ok := node.PublicKey[macAddr]
	if !ok {
		return "", errors.Errorf("public key of the user with id %s & macAddr %s does not exist. please register first", id, macAddr)
	}

	return publicKey, nil
}

func (s *SmartContract) IsNodeExists(ctx tci, id string) (bool, error) {
	nodeJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, errors.Wrap(err, "failed to read from world state")
	}

	return nodeJSON != nil, nil
}

func (s *SmartContract) GetNodeById(ctx tci, id string) (*Node, error) {
	nodeJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from world state")
	}
	if nodeJSON == nil {
		return nil, errors.New("node does not exist")
	}

	node := Node{}
	err = json.Unmarshal(nodeJSON, &node)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal error")
	}

	return &node, nil
}

func (s *SmartContract) GetAllNodes(ctx tci) ([]*Node, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var nodes []*Node

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var node Node
		err = json.Unmarshal(queryResponse.Value, &node)
		if err != nil {
			return nil, errors.Wrap(err, "json unmarshal error")
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}
