package chaincode

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"regexp"
	"time"

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
	ID                      string    `json:"ID"`
	Owner                   string    `json:"Owner"`
	AcceptanceSampling      int       `json:"AcceptanceSampling"`
	ManufacturingEquipment  int       `json:"ManufacturingEquipment"`
	TransportationEquipment int       `json:"TransportationEquipment"`
	InventoryManagement     int       `json:"InventoryManagement"`
	SaveEquipment           int       `json:"SaveEquipment"`
	A01                     int       `json:"A01"`
	A02                     int       `json:"A02"`
	A03                     int       `json:"A03"`
	A04                     int       `json:"A04"`
	B01                     int       `json:"B01"`
	C01                     int       `json:"C01"`
	Source                  string    `json:"Source"`
	TimeStamp               time.Time `json:"TimeStamp"`
	Sender                  string    `json:"Sender"`
	Function                string    `json:"Function"`
}

type Credit struct {
	ID          string  `json:"ID"`
	Transaction float32 `json:"Transaction"`
	Score       float32 `json:"Score"`
	FinalScore  float32 `json:"FinalScore"`
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
		{ID: "init", Owner: caller, AcceptanceSampling: 0, ManufacturingEquipment: 1, TransportationEquipment: 1, InventoryManagement: 0, SaveEquipment: 1, A01: 1, A02: 1, A03: 1, A04: 1, B01: 1, C01: 0, Source: caller, TimeStamp: timestamp, Sender: caller, Function: "InitLedger"},
	}

	credits := []Credit{
		{ID: "org1admin", Transaction: 0, Score: 0, FinalScore: 0},
		{ID: "org2admin", Transaction: 0, Score: 0, FinalScore: 0},
		{ID: "org5admin", Transaction: 0, Score: 0, FinalScore: 0},
		{ID: "org5admin1", Transaction: 0, Score: 0, FinalScore: 0},
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

func (s *SmartContract) GetTimeStamp(ctx contractapi.TransactionContextInterface) (time.Time, error) {
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("Failed to get timestamp")
	}

	return time.Unix(timestamp.Seconds, int64(timestamp.GetNanos())), nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, acceptancesampling int, manufacturingequipment int, transportationequipment int, inventorymanagement int, saveequipment int, a01 int, a02 int, a03 int, a04 int, b01 int, c01 int) error {
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

	var score float32 = 1
	if acceptancesampling == 0 {
		score = (score*10 - 2) / 10
	}
	if manufacturingequipment == 0 {
		score = (score*10 - 2) / 10
	}
	if transportationequipment == 0 {
		score = (score*10 - 2) / 10
	}
	if inventorymanagement == 0 {
		score = (score*10 - 2) / 10
	}
	if saveequipment == 0 {
		score = (score*10 - 2) / 10
	}

	err = s.UpdateCredit(ctx, caller, score)
	if err != nil {
		return err
	}

	score = 1
	if a01 == 0 {
		score = (score*100 - 25) / 100
	}
	if a02 == 0 {
		score = (score*100 - 25) / 100
	}
	if a03 == 0 {
		score = (score*100 - 25) / 100
	}
	if a04 == 0 {
		score = (score*100 - 25) / 100
	}
	err = s.UpdateCredit(ctx, "org2admin", score)
	if err != nil {
		return err
	}

	score = 1
	if b01 == 0 {
		score = (score*10 - 10) / 10
	}
	err = s.UpdateCredit(ctx, "org5admin", score)
	if err != nil {
		return err
	}

	score = 1
	if c01 == 0 {
		score = (score*10 - 10) / 10
	}
	err = s.UpdateCredit(ctx, "org5admin1", score)
	if err != nil {
		return err
	}

	asset := Asset{
		ID:                      id,
		Owner:                   caller,
		AcceptanceSampling:      acceptancesampling,
		ManufacturingEquipment:  manufacturingequipment,
		TransportationEquipment: transportationequipment,
		InventoryManagement:     inventorymanagement,
		SaveEquipment:           saveequipment,
		A01:                     a01,
		A02:                     a02,
		A03:                     a03,
		A04:                     a04,
		B01:                     b01,
		C01:                     c01,
		TimeStamp:               timestamp,
		Sender:                  caller,
		Function:                "CreateAsset",
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
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, acceptancesampling int, manufacturingequipment int, transportationequipment int, inventorymanagement int, saveequipment int) error {
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
	if caller != "org1admin" {
		return fmt.Errorf("only admin from org1 can use this function")
	}

	var score float32 = 1
	if acceptancesampling == 0 {
		score = (score*10 - 2) / 10
	}
	if manufacturingequipment == 0 {
		score = (score*10 - 2) / 10
	}
	if transportationequipment == 0 {
		score = (score*10 - 2) / 10
	}
	if inventorymanagement == 0 {
		score = (score*10 - 2) / 10
	}
	if saveequipment == 0 {
		score = (score*10 - 2) / 10
	}

	timestamp, err := s.GetTimeStamp(ctx)
	if err != nil {
		return err
	}

	err = s.UpdateCredit(ctx, caller, score)
	if err != nil {
		return err
	}

	// overwriting original asset with new asset
	assetNew := Asset{
		ID:                      id,
		Owner:                   caller,
		AcceptanceSampling:      acceptancesampling,
		ManufacturingEquipment:  manufacturingequipment,
		TransportationEquipment: transportationequipment,
		InventoryManagement:     inventorymanagement,
		SaveEquipment:           saveequipment,
		A01:                     asset.A01,
		A02:                     asset.A02,
		A03:                     asset.A03,
		A04:                     asset.A04,
		B01:                     asset.B01,
		C01:                     asset.C01,
		TimeStamp:               timestamp,
		Sender:                  caller,
		Function:                "Org1UpdateAsset",
	}
	assetJSON, err := json.Marshal(assetNew)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) UpdateCredit(ctx contractapi.TransactionContextInterface, id string, score float32) error {
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

	var result []*Asset
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

		result = append(result, &asset)
	}

	return result, nil
}
