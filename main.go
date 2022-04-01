package main

import (
	chaincodePkg "github.com/hiro942/elden-chain/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"log"
)

func main() {
	chaincode, err := contractapi.NewChaincode(new(chaincodePkg.SmartContract))

	if err != nil {
		log.Panicf("Error create chaincode: %v", err)
		//fmt.Printf("Error create chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
		//fmt.Printf("Error starting chaincode: %v", err)
	}
}