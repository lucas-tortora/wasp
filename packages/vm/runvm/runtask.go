package runvm

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction/txbuilder"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/packages/vm"
)

// RunComputationsAsync runs computations for the batch of requests in the background
func RunComputationsAsync(ctx *vm.VMTask) error {
	if len(ctx.Requests) == 0 {
		return fmt.Errorf("must be at least 1 request")
	}

	// create txbuilder for the task. It will accumulate all token movements produced
	// by the SC program during batch run. In the end it will produce finalized transaction
	addr := address.Address(ctx.ChainID)
	txb, err := txbuilder.NewFromAddressBalances(&addr, ctx.Balances)
	if err != nil {
		ctx.Log.Debugf("txbuilder.NewFromAddressBalances: %v\n%s", err, txutil.BalancesToString(ctx.Balances))
		return err
	}

	// TODO 1 graceful shutdown of the running VM task (with daemon)
	// TODO 2 timeout for VM. Gas limit

	go runTask(ctx, txb)

	return nil
}

// runTask runs batch of requests on VM
func runTask(ctx *vm.VMTask, txb *txbuilder.Builder) {
	ctx.Log.Debugw("runTask IN",
		"chainID", ctx.ChainID.String(),
		"timestamp", ctx.Timestamp,
		"state index", ctx.VirtualState.StateIndex(),
		"num req", len(ctx.Requests),
		"leader", ctx.LeaderPeerIndex,
	)

	// create VM context, including state block, move smart contract token and request tokens
	vmctx, err := createVMContext(ctx, txb)
	if err != nil {
		ctx.OnFinish(fmt.Errorf("runTask.createVMContext: %v", err))
		return
	}
	stateUpdates := make([]state.StateUpdate, 0, len(ctx.Requests))
	for _, reqRef := range ctx.Requests {

		vmctx.RequestRef = reqRef
		vmctx.StateUpdate = state.NewStateUpdate(reqRef.RequestID()).WithTimestamp(vmctx.Timestamp)

		runTheRequest(vmctx)

		stateUpdates = append(stateUpdates, vmctx.StateUpdate)
		// update state
		vmctx.VirtualState.ApplyStateUpdate(vmctx.StateUpdate)
		if vmctx.Timestamp != 0 {
			// increasing (nonempty) timestamp for 1 nanosecond for each request in the batch
			// the reason is to provide a different timestamp for each VM call and remain deterministic
			vmctx.Timestamp += 1
		}
		// mutate entropy
		vmctx.Entropy = *hashing.HashData(vmctx.Entropy[:])
	}
	if len(stateUpdates) == 0 {
		// should not happen
		ctx.OnFinish(fmt.Errorf("RunVM: no state updates were produced"))
		return
	}

	// create batch from state updates.
	ctx.ResultBlock, err = state.NewBlock(stateUpdates)
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.NewBlock: %v", err))
		return
	}
	ctx.ResultBlock.WithStateIndex(ctx.VirtualState.StateIndex() + 1)

	// calculate resulting state hash
	vsClone := ctx.VirtualState.Clone()
	if err = vsClone.ApplyBatch(ctx.ResultBlock); err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.ApplyBatch: %v", err))
		return
	}
	stateHash := vsClone.Hash()

	// add state block
	err = vmctx.TxBuilder.SetStateParams(ctx.VirtualState.StateIndex()+1, stateHash, vsClone.Timestamp())
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.txbuilder.SetStateParams: %v", err))
		return
	}
	// create result transaction
	ctx.ResultTransaction, err = vmctx.TxBuilder.Build(false)
	if err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.txbuilder.Build: %v", err))
		return
	}
	// check semantic just in case
	if _, err := ctx.ResultTransaction.Properties(); err != nil {
		ctx.OnFinish(fmt.Errorf("RunVM.txbuilder.Properties: %v", err))
		return
	}

	ctx.Log.Debugw("runTask OUT",
		"result batch size", ctx.ResultBlock.Size(),
		"result batch state index", ctx.ResultBlock.StateIndex(),
		"result variable state hash", stateHash.String(),
		"result essence hash", hashing.HashData(ctx.ResultTransaction.EssenceBytes()).String(),
		"result tx finalTimestamp", time.Unix(0, ctx.ResultTransaction.MustState().Timestamp()),
	)
	// call back
	ctx.OnFinish(nil)
}
