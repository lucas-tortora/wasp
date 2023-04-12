// Code generated by schema tool; DO NOT EDIT.

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

//nolint:dupl
package solotutorialimpl

import (
	"github.com/iotaledger/wasp/documentation/tutorial-examples/go/solotutorial"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

var exportMap = wasmlib.ScExportMap{
	Names: []string{
		solotutorial.FuncStoreString,
		solotutorial.ViewGetString,
	},
	Funcs: []wasmlib.ScFuncContextFunction{
		funcStoreStringThunk,
	},
	Views: []wasmlib.ScViewContextFunction{
		viewGetStringThunk,
	},
}

func OnDispatch(index int32) *wasmlib.ScExportMap {
	return exportMap.Dispatch(index)
}

type StoreStringContext struct {
	Params solotutorial.ImmutableStoreStringParams
	State  solotutorial.MutableSoloTutorialState
}

func funcStoreStringThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("solotutorial.funcStoreString")
	f := &StoreStringContext{
		Params: solotutorial.NewImmutableStoreStringParams(),
		State:  solotutorial.NewMutableSoloTutorialState(),
	}
	ctx.Require(f.Params.Str().Exists(), "missing mandatory param: str")
	funcStoreString(ctx, f)
	ctx.Log("solotutorial.funcStoreString ok")
}

type GetStringContext struct {
	Results solotutorial.MutableGetStringResults
	State   solotutorial.ImmutableSoloTutorialState
}

func viewGetStringThunk(ctx wasmlib.ScViewContext) {
	ctx.Log("solotutorial.viewGetString")
	f := &GetStringContext{
		Results: solotutorial.NewMutableGetStringResults(),
		State:   solotutorial.NewImmutableSoloTutorialState(),
	}
	viewGetString(ctx, f)
	ctx.Results(f.Results.Proxy)
	ctx.Log("solotutorial.viewGetString ok")
}
