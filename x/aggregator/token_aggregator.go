package aggregator

import (
	"github.com/bluzelle/curium/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"regexp"
	"strings"
)

var AggValueKey = types.NewOracleStoreKey(0x10)

type TokenAggregator struct {
}

type AggValue struct {
	Batch    string
	Symbol   string
	InSymbol string
	Value    sdk.Dec
	Count    int64
	Height   int64
}

type AggSourceValue struct {
	SourceName string
	Symbol string
	InSymbol string
	Batch string
	Value sdk.Dec
	Owner sdk.AccAddress
	Height int64
	Count int64
	Weight int64
}

type QueryReqSearchBatches struct {
	PreviousBatch string
	Reverse bool
	Limit uint
}


func convertToAggSourceValues(values []types.SourceValue) []AggSourceValue {
	var results []AggSourceValue
	for _, value := range(values) {
		parts := strings.Split(value.SourceName, "-")
		results = append(results, AggSourceValue{
			SourceName: value.SourceName,
			Symbol:     parts[1],
			InSymbol:   parts[3],
			Batch:      value.Batch,
			Value:      value.Value,
			Owner:      value.Owner,
			Height:     value.Height,
			Count:      value.Count,
			Weight:     value.Weight,
		})
	}
	return results
}

func (ta TokenAggregator) GetName() string {
	return "token-price"
}

func (ta TokenAggregator) Queriers(ctx sdk.Context, cmd string, req abci.RequestQuery, cdc codec.Codec, store sdk.KVStore) (bool, []byte, error) {
	switch cmd {
	case "searchBatchKeys":
		resp, err := querySearchBatchKeys(ctx, req, cdc, store)
		return true, resp, err
	}
	return false, nil, nil
}

func querySearchBatchKeys(ctx sdk.Context, req abci.RequestQuery, cdc codec.Codec, store sdk.KVStore) ([]byte, error) {
	var query QueryReqSearchBatches
	cdc.MustUnmarshalJSON(req.Data, &query)

	result := searchAggValueBatchKeys(ctx, store, query.PreviousBatch, query.Limit, query.Reverse)
	return cdc.MustMarshalJSON(result), nil
}


func (ta TokenAggregator) AggregateSourceValues(ctx sdk.Context, cdc codec.Codec, store sdk.KVStore, sourceValues []types.SourceValue) {
	values := convertToAggSourceValues(sourceValues)
	fixupUsdItems(values)
	values = addInverses(values)


	averagers := make(map[string]Averager)

	for _, value := range values {
		key := value.Symbol + "-" + value.InSymbol
		x := averagers[key]
		x.Add(value.Value, value.Weight)
		averagers[key] = x
	}

	for key, averager := range averagers {
		parts := strings.Split(key, "-")

		storeKey := AggValueKey.MakeKey(values[0].Batch, parts[0], parts[1])

		value := AggValue{
			Batch:    values[0].Batch,
			Symbol:   parts[0],
			InSymbol: parts[1],
			Value:    averager.CalculateAverage(),
			Count:    averager.Count,
			Height:		values[0].Height,
		}
		store.Set(storeKey, cdc.MustMarshalBinaryBare(value))
	}

}

func fixupUsdItems(values []AggSourceValue) {
	for i, value := range values {
		if value.Symbol == "usdt" || value.Symbol == "usdc" {
			values[i].Symbol = "usd"
		}
		if value.InSymbol == "usdt" || value.InSymbol == "usdc" {
			values[i].InSymbol = "usd"
		}
	}
}

func addInverses(values []AggSourceValue) []AggSourceValue {
	var inverses []AggSourceValue
	for _, value := range values {
		inverses = append(inverses, AggSourceValue{
			SourceName: value.SourceName + "-inverse",
			Batch:      value.Batch,
			Symbol:     value.InSymbol,
			InSymbol:   value.Symbol,
			Value:      sdk.NewDec(1).Quo(value.Value),
			Weight:		value.Weight,
		})
	}
	return append(values, inverses...)
}


func searchAggValueBatchKeys(ctx sdk.Context, store sdk.KVStore, previousKey string, limit uint, reverse bool) []string {
	isInRange := func(key string) bool {
		if len(previousKey) == 0 {
			return true
		}

		if reverse && key < previousKey {
			return true
		}
		if !reverse && key > previousKey {
			return true
		}
		return false
	}

	var batches = make([]string, 0)

	if limit == 0 {
		limit = 50
	}

	var iterator sdk.Iterator

	if reverse {
		iterator = sdk.KVStoreReversePrefixIterator(store, AggValueKey.Prefix)
	} else {
		iterator = store.Iterator(nil, nil)
	}
	defer iterator.Close()

	var re = regexp.MustCompile(`([^>]*).*`)
	var prevBatch string

	for ;iterator.Valid(); iterator.Next() {
		key := string(iterator.Key()[1:])
		batch := re.ReplaceAllString(key, `$1`)

		if isInRange(batch) {

			if batch != prevBatch {
				batches = append(batches, batch)
			}
			prevBatch = batch
		}

		if uint(len(batches)) == limit {
			break
		}
	}

	return batches
}



