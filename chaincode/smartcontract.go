package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Component struct {
	Category string `json:"category"`
	Hash     string `json:"hash"`
}

type Asset struct {
	Nodes         []Component `json:"nodes"`
	Relationships []Component `json:"relationships"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record Asset
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, data string) error {

	return ctx.GetStub().PutState(assetID, []byte(data))
}

func (s *SmartContract) GetAsset(ctx contractapi.TransactionContextInterface, assetID string) ([]byte, error) {
	assetAsBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from Ledger. %s", err.Error())
	}
	if assetAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", assetID)
	}
	return assetAsBytes, nil
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}
		block := new(Asset)
		_ = json.Unmarshal(queryResponse.Value, &block)
		queryResult := QueryResult{
			Key:    queryResponse.Key,
			Record: *block,
		}
		results = append(results, queryResult)
	}
	return results, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
