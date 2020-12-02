package viewcontext

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/hardcoded"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type viewcontext struct {
	processors *processors.ProcessorCache
	state      buffered.BufferedKVStore
	chainID    coretypes.ChainID
}

func New(chain chain.Chain) (*viewcontext, error) {
	state, _, ok, err := state.LoadSolidState(chain.ID())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("State not found for chain %s", chain.ID())
	}
	return &viewcontext{
		processors: chain.Processors(),
		state:      state.Variables(),
		chainID:    *chain.ID(),
	}, nil
}

func (v *viewcontext) CallView(contractHname coretypes.Hname, epCode coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	rec, err := root.FindContract(contractStateSubpartition(v.state, root.Interface.Hname()), contractHname)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract %s: %v", contractHname, err)
	}

	proc, err := v.processors.GetOrCreateProcessor(rec, func(programHash hashing.HashValue) (string, []byte, error) {
		if vmtype, ok := hardcoded.LocateHardcodedProgram(programHash); ok {
			return vmtype, nil, nil
		}
		return blob.LocateProgram(contractStateSubpartition(v.state, blob.Interface.Hname()), programHash)
	})
	if err != nil {
		return nil, err
	}

	ep, ok := proc.GetEntryPoint(epCode)
	if !ok {
		return nil, fmt.Errorf("%s: can't find entry point '%s'", proc.GetDescription(), epCode.String())
	}

	if !ep.IsView() {
		return nil, fmt.Errorf("only view entry point can be called in this context")
	}

	return ep.CallView(newSandboxView(v, coretypes.NewContractID(v.chainID, contractHname), params))
}

func contractStateSubpartition(state kv.KVStore, contractHname coretypes.Hname) kv.KVStore {
	return subrealm.New(state, kv.Key(contractHname.Bytes()))
}
