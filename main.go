package main

import (
	chaincodePkg "github.com/hiro942/elden-chain/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"time"
)

func main() {
	time.Local = time.FixedZone("CST", 8*3600)

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