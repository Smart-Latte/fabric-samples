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

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Energy struct {
	AppraisedValue int       `json:"AppraisedValue"`
	GeneratedTime  time.Time `json:"Generated Time"`
	PurchasedTime  time.Time `json:"Purchased Time"`
	ID             string    `json:"ID"`
	LargeCategory  string    `json:"LargeCategory"`
	Latitude       float64   `json:"Latitude"`
	Longitude      float64   `json:"Longitude"`
	Owner          string    `json:"Owner"`
	Producer       string    `json:"Producer"`
	SmallCategory  string    `json:"SmallCategory"`
	Status         string    `json:"Status"`
}

// InitLedger adds a base set of assets to the ledger
// Owner: Brad, Jin Soo, Max, Adriana, Michel
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	energies := []Energy{
		{ID: "energy1", LargeCategory: "Green", SmallCategory: "solor", Status: "generated",
			Producer: "Tomoko", Latitude: 1, Longitude: 1, AppraisedValue: 0},
		{ID: "energy2", LargeCategory: "Green", SmallCategory: "solor", Status: "generated",
			Producer: "Tomoko", Owner: "Brad", Latitude: 1, Longitude: 1, AppraisedValue: 0},
		{ID: "energy3", LargeCategory: "Green", SmallCategory: "solor", Status: "sold",
			Owner: "Jin Soo", Latitude: 1, Longitude: 1, AppraisedValue: 100},
		{ID: "energy4", LargeCategory: "Green", SmallCategory: "solor", Status: "generated",
			Owner: "Max", Latitude: 1, Longitude: 1, AppraisedValue: 0},
		{ID: "energy5", LargeCategory: "Green", SmallCategory: "solor", Status: "generated",
			Owner: "Adriana", Latitude: 1, Longitude: 1, AppraisedValue: 0},
		{ID: "energy6", LargeCategory: "Green", SmallCategory: "solor", Status: "generated",
			Owner: "Michel", Latitude: 1, Longitude: 1, AppraisedValue: 0},
	}

	for _, energy := range energies {
		energyJSON, err := json.Marshal(energy)
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

// CreateAsset issues a new asset to the world state with given details.
// 新しいトークンの発行
// errorは返り値の型
// 引数は、ID、緯度、経度、エネルギーの種類、発電した時間、発電者、価格
// トークンには、オーナー、ステータスも含める
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface,
	id string, latitude float64, longitude float64, producer string, largeCategory string, smallCategory string, appraisedValue int) error {
	exists, err := s.EnergyExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the energy %s already exists", id)
	}

	energy := Energy{
		ID:             id,
		Latitude:       latitude,
		Longitude:      longitude,
		Owner:          producer,
		Producer:       producer,
		LargeCategory:  largeCategory,
		SmallCategory:  smallCategory,
		AppraisedValue: appraisedValue,
		Status:         "generated",
		GeneratedTime:  time.Now(),
	}
	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, energyJSON)
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
// 購入する
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	energy, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	//generatedTime := energy.GeneratedTime
	var purchasedTime = time.Now()
	var tCompare = purchasedTime.Add(time.Minute * -30)

	// UpdateAssetを改良してステータスを変更するようにしても良いかも
	if energy.GeneratedTime.After(tCompare) == true {
		energy.PurchasedTime = time.Now()
	} else {
		return "", fmt.Errorf("the energy %s was generated more than 30min ago", id)
	}

	if energy.Status == "generated" {
		energy.Status = "sold"
	} else {
		return "", fmt.Errorf("the energy %s is not for sale", id)
	}

	oldOwner := energy.Owner
	energy.Owner = newOwner

	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, energyJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

// AssetExists returns true when asset with given ID exists in world state
// スタブの意味はよく分からない。台帳にアクセスするための関数らしい。一般的には「外部プログラムとの細かなインターフェース制御を引き受けるプログラム」を指すらしい
func (s *SmartContract) EnergyExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	energyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return energyJSON != nil, nil
}

// ReadAsset returns the asset stored in the world state with given id.
// アセットを返す
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Energy, error) {
	energyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if energyJSON == nil {
		return nil, fmt.Errorf("the energy %s does not exist", id)
	}

	var energy Energy
	err = json.Unmarshal(energyJSON, &energy)
	if err != nil {
		return nil, err
	}

	return &energy, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
// 内容は読み込まない。存在することを確認し、上書きする
// ステータスを変更するため改造
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.EnergyExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the energy %s does not exist", id)
	}

	// overwriting original asset with new asset
	energy := Energy{
		ID:     id,
		Status: "old",
	}
	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, energyJSON)
}

func (s *SmartContract) QueryByStatus(ctx contractapi.TransactionContextInterface, status string) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"status":"%s"}}`, status)
	// queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var energies []*Energy
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var energy Energy
		err = json.Unmarshal(queryResponse.Value, &energy)
		if err != nil {
			return nil, err
		}
		energies = append(energies, &energy)
	}

	return energies, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Energy, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var energies []*Energy
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var energy Energy
		err = json.Unmarshal(queryResponse.Value, &energy)
		if err != nil {
			return nil, err
		}
		energies = append(energies, &energy)
	}

	return energies, nil
}

// DeleteAsset deletes an given asset from the world state.
// 後回し？そもそもいらない？30分経ったものを消去するかどうか。ステータスを変更するにとどめるか
// 現状はUpdateAssetでステータスを変更
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.EnergyExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the energy %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
