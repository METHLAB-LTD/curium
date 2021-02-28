package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/bluzelle/curium/x/oracle/types"
	"time"
	storeIterator "github.com/cosmos/cosmos-sdk/store/types"
)

func (k Keeper) RegisterValueUpdatedListener(listener types.ValueUpdateListener) {
	valueUpdateListeners = append(valueUpdateListeners, listener)
}

func (k Keeper) GetValueStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.valueStoreKey)
}

func (k Keeper) UpdateSourceValue(ctx sdk.Context, vote types.Vote) {
	votes := k.SearchVotes(ctx, makeSearchVotePrefix(vote.Batch, vote.SourceName))
	votes = append(votes, vote)
	average := calculateAverageFromVotes(votes)
	store := k.GetValueStore(ctx)
	key := MakeSourceValueKey(vote.Batch, vote.SourceName)
	sourceValue := types.SourceValue{
		SourceName: vote.SourceName,
		Batch:      vote.Batch,
		Value:      average,
		Owner:      vote.Owner,
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(sourceValue))

	for _, listener := range valueUpdateListeners {
		// Delay the execution of the listener so it happens after the block is committed
		time.AfterFunc(time.Second, func() {listener(ctx, sourceValue)})
	}
}

func MakeSourceValueKey(batch string, sourceName string) string {
	return fmt.Sprintf("%s>%s", batch, sourceName)
}


func calculateAverageFromVotes(votes []types.Vote) sdk.Dec {
	sum := sdk.NewDecFromInt(sdk.NewInt(0))
	totalWeight := sdk.NewDecFromInt(sdk.NewInt(0))

	for _, vote := range votes  {
		totalWeight = totalWeight.Add(vote.Weight)
		sum = sum.Add(vote.Value.Mul(vote.Weight))
	}

	return sum.Quo(totalWeight)
}

func (k Keeper) SearchSourceValues(ctx sdk.Context, prefix string, page uint, limit uint, reverse bool) []types.SourceValue {
	var iterator sdk.Iterator
	if reverse {
		iterator = storeIterator.KVStoreReversePrefixIteratorPaginated(k.GetValueStore(ctx), []byte(prefix), page, limit)
	} else {
		iterator = storeIterator.KVStorePrefixIteratorPaginated(k.GetValueStore(ctx), []byte(prefix), page, limit)
	}
	defer iterator.Close()
	values  := make([]types.SourceValue, 0)

	for ;iterator.Valid(); iterator.Next() {
		if ctx.GasMeter().IsPastLimit() {
			break
		}

		var v types.SourceValue
		value := iterator.Value()
		k.cdc.MustUnmarshalBinaryBare(value, &v)
		values = append(values, v)
	}
	return values
}



