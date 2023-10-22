package chaincode

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"regexp"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	AppraisedValue int    `json:"AppraisedValue"`
	Color          string `json:"Color"`
	ID             string `json:"ID"`
	Owner          string `json:"Owner"`
	Size           int    `json:"Size"`
	Source         string `json:"Source"`
	TimeStamp      string `json:"TimeStamp"`
	Sender         string `json:"Sender"`
	Function       string `json:"Function"`
}

type Credit struct {
	ID          string `json:"ID"`
	Transaction int    `json:"Transaction"`
	Score       int    `json:"Score"`
	FinalScore  int    `json:"FinalScore"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	timestamp, err := s.GetTimeStamp(ctx)
	if err != nil {
		return err
	}

	caller, err := s.GetCallerName(ctx)
	if err != nil {
		return err
	}

	assets := []Asset{
		{ID: "asset1", Color: "blue", Size: 5, Owner: caller, AppraisedValue: 300, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
		{ID: "asset2", Color: "red", Size: 5, Owner: caller, AppraisedValue: 400, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
		{ID: "asset3", Color: "green", Size: 10, Owner: caller, AppraisedValue: 500, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: caller, AppraisedValue: 600, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
		{ID: "asset5", Color: "black", Size: 15, Owner: caller, AppraisedValue: 700, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
		{ID: "asset6", Color: "white", Size: 15, Owner: caller, AppraisedValue: 800, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
	}

	credits := []Credit{
		{ID: "org1admin", Transaction: 0, Score: 0, FinalScore: 0},
		{ID: "org2admin", Transaction: 0, Score: 0, FinalScore: 0},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, credit := range credits {
		creditJSON, err := json.Marshal(credit)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(credit.ID, creditJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// GetCallerName get the caller's identity
func (s *SmartContract) GetCallerName(ctx contractapi.TransactionContextInterface) (string, error) {
	caller, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", fmt.Errorf("failed to get the caller's identity")
	}
	re := regexp.MustCompile("-----BEGIN CERTIFICATE-----[^ ]+-----END CERTIFICATE-----\n")
	match := re.FindStringSubmatch(string(caller))
	pemBlock, _ := pem.Decode([]byte(match[0]))
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return "", err
	}
	str := cert.Subject.CommonName

	return str, nil
}

func (s *SmartContract) GetTimeStamp(ctx contractapi.TransactionContextInterface) (string, error) {
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("Failed to get timestamp")
	}

	return timestamp.String(), err
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	caller, err := s.GetCallerName(ctx)
	if err != nil {
		return err
	}

	if caller != "org1admin" {
		return fmt.Errorf("only admin from org1 can use this method")
	}

	timestamp, err := s.GetTimeStamp(ctx)
	if err != nil {
		return err
	}

	var score = 1
	if color == "" {
		score = (score*10 - 3) / 10
	}
	if size == 0 {
		score = (score*10 - 3) / 10
	}
	if appraisedValue == 0 {
		score = (score*10 - 4) / 10
	}

	err = s.UpdateCredit(ctx, caller, score)
	if err != nil {
		return err
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          caller,
		AppraisedValue: appraisedValue,
		TimeStamp:      timestamp,
		Sender:         caller,
		Function:       "CreateAsset",
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *SmartContract) ReadCredit(ctx contractapi.TransactionContextInterface, id string) (*Credit, error) {
	creditJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if creditJSON == nil {
		return nil, fmt.Errorf("the credit %s does not exist", id)
	}

	var credit Credit
	err = json.Unmarshal(creditJSON, &credit)
	if err != nil {
		return nil, err
	}

	return &credit, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	caller, err := s.GetCallerName(ctx)
	if err != nil {
		return err
	}

	if asset.Owner != caller {
		return fmt.Errorf("only the owner of the asset can update the asset")
	}

	timestamp, err := s.GetTimeStamp(ctx)
	if err != nil {
		return err
	}

	var score = 1
	if color == "" {
		score = (score*10 - 3) / 10
	}
	if size == 0 {
		score = (score*10 - 3) / 10
	}
	if appraisedValue == 0 {
		score = (score*10 - 4) / 10
	}

	err = s.UpdateCredit(ctx, caller, score)
	if err != nil {
		return err
	}

	// overwriting original asset with new asset
	assetNew := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          caller,
		AppraisedValue: appraisedValue,
		TimeStamp:      timestamp,
		Sender:         caller,
		Function:       "UpdateAsset",
	}
	assetJSON, err := json.Marshal(assetNew)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) UpdateCredit(ctx contractapi.TransactionContextInterface, id string, score int) error {
	credit, err := s.ReadCredit(ctx, id)
	if err != nil {
		return err
	}

	credit.Score += score
	credit.Transaction += 1
	credit.FinalScore = credit.Score / credit.Transaction

	creditJSON, err := json.Marshal(credit)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, creditJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	caller, err := s.GetCallerName(ctx)
	if err != nil {
		return err
	}

	if asset.Owner != caller {
		return fmt.Errorf("only the owner of the asset can delete the asset")
	}

	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	caller, err := s.GetCallerName(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot get the caller's identity")
	}

	oldOwner := asset.Owner

	if caller != oldOwner {
		return "", fmt.Errorf("only the owner of this assets can do the transfer, now the owner is: %v", caller)
	}

	timestamp, err := s.GetTimeStamp(ctx)
	if err != nil {
		return "", err
	}

	asset.Owner = newOwner
	asset.Function = "TransferAsset"
	asset.Sender = caller
	asset.TimeStamp = timestamp

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

func (s *SmartContract) GetHistoryForKey(ctx contractapi.TransactionContextInterface, id string) ([]*Asset, error) {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	assetJSON, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get history data for asset: %s", id)
	}
	defer assetJSON.Close()

	var assets []*Asset
	for assetJSON.HasNext() {
		queryResponse, err := assetJSON.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
