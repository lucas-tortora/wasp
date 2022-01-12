// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package inccounter

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

type ImmutableIncrementWithDelayParams struct {
	id int32
}

func (s ImmutableIncrementWithDelayParams) Delay() wasmlib.ScImmutableInt32 {
	return wasmlib.NewScImmutableInt32(s.id, wasmlib.KeyID(ParamDelay))
}

type MutableIncrementWithDelayParams struct {
	id int32
}

func (s MutableIncrementWithDelayParams) Delay() wasmlib.ScMutableInt32 {
	return wasmlib.NewScMutableInt32(s.id, wasmlib.KeyID(ParamDelay))
}

type ImmutableInitParams struct {
	id int32
}

func (s ImmutableInitParams) Counter() wasmlib.ScImmutableInt64 {
	return wasmlib.NewScImmutableInt64(s.id, idxMap[IdxParamCounter])
}

type MutableInitParams struct {
	id int32
}

func (s MutableInitParams) Counter() wasmlib.ScMutableInt64 {
	return wasmlib.NewScMutableInt64(s.id, idxMap[IdxParamCounter])
}

type ImmutableRepeatManyParams struct {
	id int32
}

func (s ImmutableRepeatManyParams) NumRepeats() wasmlib.ScImmutableInt64 {
	return wasmlib.NewScImmutableInt64(s.id, wasmlib.KeyID(ParamNumRepeats))
}

type MutableRepeatManyParams struct {
	id int32
}

func (s MutableRepeatManyParams) NumRepeats() wasmlib.ScMutableInt64 {
	return wasmlib.NewScMutableInt64(s.id, wasmlib.KeyID(ParamNumRepeats))
}

type ImmutableWhenMustIncrementParams struct {
	id int32
}

func (s ImmutableWhenMustIncrementParams) Dummy() wasmlib.ScImmutableInt64 {
	return wasmlib.NewScImmutableInt64(s.id, wasmlib.KeyID(ParamDummy))
}

type MutableWhenMustIncrementParams struct {
	id int32
}

func (s MutableWhenMustIncrementParams) Dummy() wasmlib.ScMutableInt64 {
	return wasmlib.NewScMutableInt64(s.id, wasmlib.KeyID(ParamDummy))
}
