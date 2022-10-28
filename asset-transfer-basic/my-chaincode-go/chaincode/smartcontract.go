package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Energyは単純なアセットを構成する基本的な詳細を記述する
//構造体フィールドをアルファベット順に挿入 => 言語間の決定論を実現するため
// golangではjsonにmarshalする際に順番を保持するが、自動的には並べない
type Energy struct {
	AppraisedValue    int    `json:"AppraisedValue"`
	GeneratedTime    time.Time `json:"Generated Time`
	PurchasedTime     time.Time `json:"Purchased Time`
	ID                string `json:"ID"`
	Latitude          float64 `json:"Latitude"`
	Longitude         float64 `json:"Longitude"`
	Owner             string `json:"Owner"`
	Producer          string `json:"Producer`
	Status            string `json:"Status"`
	Type              string `json:"Type"`
}

// InitLedgerは台帳に基本的な資産のセットを追加
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var t1 = time.Now().Add(time.Hour * -1)
	var t2 = time.Now().Add(time.Minute * -50)
	var t3 = time.Now().Add(time.Minute * -40)
	var t4 = time.Now().Add(time.Minute * -30)
	var t5 = time.Now().Add(time.Minute * -20)
	var t6 = time.Now()

	energys := []Energy{
		{ID: "energy1", Type: "solor", Status: "generated", Producer: "Tomoko", Latitude: 1, Longitude: 1, GeneratedTime: t1, AppraisedValue: 0},
		{ID: "energy2", Type: "solor", Status: "generated", Producer: "Tomoko", Owner: "Brad", Latitude: 1, GeneratedTime: t2, PurchasedTime: t3, Longitude: 1, AppraisedValue: 0},
		{ID: "energy3", Type: "solor", Status: "sold", Owner: "Jin Soo", Latitude: 1, Longitude: 1, GeneratedTime: t3, AppraisedValue: 100},
		{ID: "energy4", Type: "solor", Status: "generated", Owner: "Max", Latitude: 1, Longitude: 1, GeneratedTime: t4, AppraisedValue: 0},
		{ID: "energy5", Type: "solor", Status: "generated", Owner: "Adriana", Latitude: 1, Longitude: 1, GeneratedTime: t5, AppraisedValue: 0},
		{ID: "energy6", Type: "solor", Status: "generated", Owner: "Michel", Latitude: 1, Longitude: 1, GeneratedTime: t6, AppraisedValue: 0},
	}

	for _, energy := range energys {
		energyJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(energy.ID, energyJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAssetは与えられた詳細を持つ新しいアセットをthe world stateに発行
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	let results = await query('selectAssets');
	let count = results.length;
	
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
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

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
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

	oldOwner := asset.Owner
	asset.Owner = newOwner

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
