package synchronizer

import (
	"context"
	"fmt"
	"github.com/bluzelle/curium/x/curium"
	"github.com/bluzelle/curium/x/synchronizer/contract"
	"github.com/bluzelle/curium/x/synchronizer/keeper"
	"github.com/bluzelle/curium/x/synchronizer/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	eth "github.com/ethereum/go-ethereum/ethclient"
	"sync"
	"time"
)

const CONTRACT_ADDRESS = "0x4Fe0D5763cF500454E2b105f6AE8b9b66Ea4dD64"

var doOnce sync.Once
var currCtx sdk.Context
var ticker = time.NewTicker(time.Second * 15)

func StartSynchronizer(ctx sdk.Context, k keeper.Keeper) {
	currCtx = ctx
	doOnce.Do(func() {
		go func() {
			for {
				_ = <-ticker.C
				runSynchronizer(k)
			}
		}()
	})
}

func runSynchronizer(k keeper.Keeper) {
	sources := k.GetAllSource(currCtx)
	var voteMessages []sdk.Msg

	for _, source := range sources {
		data := fetchDataFromContract(source)
		for _, item := range data {
			msg, err := generateVoteMsg(source, item, k)
			voteMessages = append(voteMessages, msg)
			if err != nil {
				k.Logger(currCtx).Info("Error creating vote message", "item", item)
			}
		}
	}
	curium.BroadcastMessages(currCtx, voteMessages, k.AccKeeper, "sync", k.GetKeyringDir())
}

func fetchDataFromContract(source types.Source) []contract.TestingRecord {
	ethCtx := context.Background()
	backend, err := eth.Dial(source.Url)
	if err != nil {
		fmt.Println(err)
	}

	addr := common.HexToAddress(CONTRACT_ADDRESS)
	ctr, err := contract.NewTesting(addr, backend)
	if err != nil {
		fmt.Println(err)
	}
	callOpts := &bind.CallOpts{Context: ethCtx, Pending: false}
	data, err := ctr.GetSynchronizerData(callOpts)
	return data
}

func generateVoteMsg(source types.Source, record contract.TestingRecord, k keeper.Keeper) (sdk.Msg, error) {
	kr, err := keyring.New("curium", keyring.BackendTest, k.GetKeyringDir(), nil)

	if err != nil {
		return nil, err
	}
	key, err := kr.Key("sync")
	if err != nil {
		return nil, err
	}

	voteMsg := types.NewMsgSynchronizerVote(
		key.GetAddress().String(),
		record.Opt,
		record.Key,
		record.Value,
		)
	err = voteMsg.ValidateBasic()
	if err != nil {
		k.Logger(currCtx).Info("Error generating vote message", "source", source)
		return nil, err
	}
	return voteMsg, nil
}

