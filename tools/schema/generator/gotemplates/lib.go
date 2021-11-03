package gotemplates

var libGo = map[string]string{
	// *******************************
	"lib.go": `
//nolint:dupl
$#emit goHeader

func OnLoad() {
	exports := wasmlib.NewScExports()
$#each func libExportFunc

	for i, key := range keyMap {
		idxMap[i] = key.KeyID()
	}
}
$#each func libThunk
`,
	// *******************************
	"libExportFunc": `
	exports.Add$Kind($Kind$FuncName, $kind$FuncName$+Thunk)
`,
	// *******************************
	"libThunk": `

type $FuncName$+Context struct {
$#if param ImmutableFuncNameParams
$#if result MutableFuncNameResults
$#if func MutablePackageState
$#if view ImmutablePackageState
}

func $kind$FuncName$+Thunk(ctx wasmlib.Sc$Kind$+Context) {
	ctx.Log("$package.$kind$FuncName")
$#func accessCheck
	f := &$FuncName$+Context{
$#if param ImmutableFuncNameParamsInit
$#if result MutableFuncNameResultsInit
$#if func MutablePackageStateInit
$#if view ImmutablePackageStateInit
	}
$#each mandatory requireMandatory
	$kind$FuncName(ctx, f)
	ctx.Log("$package.$kind$FuncName ok")
}
`,
	// *******************************
	"ImmutableFuncNameParams": `
	Params  Immutable$FuncName$+Params
`,
	// *******************************
	"ImmutableFuncNameParamsInit": `
		Params: Immutable$FuncName$+Params{
			id: wasmlib.OBJ_ID_PARAMS,
		},
`,
	// *******************************
	"MutableFuncNameResults": `
	Results Mutable$FuncName$+Results
`,
	// *******************************
	"MutableFuncNameResultsInit": `
		Results: Mutable$FuncName$+Results{
			id: wasmlib.OBJ_ID_RESULTS,
		},
`,
	// *******************************
	"MutablePackageState": `
	State   Mutable$Package$+State
`,
	// *******************************
	"MutablePackageStateInit": `
		State: Mutable$Package$+State{
			id: wasmlib.OBJ_ID_STATE,
		},
`,
	// *******************************
	"ImmutablePackageState": `
	State   Immutable$Package$+State
`,
	// *******************************
	"ImmutablePackageStateInit": `
		State: Immutable$Package$+State{
			id: wasmlib.OBJ_ID_STATE,
		},
`,
	// *******************************
	"requireMandatory": `
	ctx.Require(f.Params.$FldName().Exists(), "missing mandatory $fldName")
`,
	// *******************************
	"grantForKey": `
	access := ctx.State().GetAgentID(wasmlib.Key("$grant"))
	ctx.Require(access.Exists(), "access not set: $grant")
`,
	// *******************************
	"grantRequire": `
	ctx.Require(ctx.Caller() == $grant, "no permission")

`,
}

