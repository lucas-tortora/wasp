package generator

var templates = map[string]string{
	// *******************************
	"copyright": `
// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
`,
	// *******************************
	"warning": `

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

`,
}
