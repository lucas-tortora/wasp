package micropay

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"time"
)

func initialize(_ coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

func publicKey(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := coreutil.NewAssert(ctx.Log())
	a.Require(ctx.Caller().IsAddress(), "micropay.publicKey: caller must be an address")

	par := kvdecoder.New(ctx.Params(), ctx.Log())

	pubKeyBin := par.MustGetBytes(ParamPublicKey)
	pubKey, _, err := ed25519.PublicKeyFromBytes(pubKeyBin)
	a.RequireNoError(err)
	addr := address.FromED25519PubKey(pubKey)
	a.Require(addr == ctx.Caller().MustAddress(), "public key does not correspond to the caller's address")

	pkRegistry := collections.NewMap(ctx.State(), StateVarPublicKeys)
	a.RequireNoError(pkRegistry.SetAt(addr[:], pubKeyBin))
	return nil, nil
}

// addWarrant adds payment warrant for specific service address
// Params:
// - ParamServiceAddress address.Address
func addWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := coreutil.NewAssert(ctx.Log())

	a.Require(ctx.Caller().IsAddress(), "payer must be an address")
	payerAddr := ctx.Caller().MustAddress()

	a.Require(getPublicKey(ctx.State(), payerAddr, a) != nil,
		fmt.Sprintf("unknown public key for address %s", payerAddr))

	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	addWarrant := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	a.Require(addWarrant >= MinimumWarrantIotas, fmt.Sprintf("warrant must be larger than %d iotas", MinimumWarrantIotas))

	warrant, revoke := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(revoke == 0, fmt.Sprintf("warrant of %s for %s is being revoked", payerAddr, serviceAddr))

	setWarrant(ctx.State(), payerAddr, serviceAddr, warrant+addWarrant)

	// all non-iota token accrue on-chain to the caller
	sendBack := ctx.IncomingTransfer().TakeOutColor(balance.ColorIOTA)
	err := accounts.Accrue(ctx, ctx.Caller(), sendBack)
	a.RequireNoError(err)

	ctx.Event(fmt.Sprintf("[micropay.addWarrant] %s increased warrant %d -> %d i for %s",
		payerAddr, warrant, warrant+addWarrant, serviceAddr))
	return nil, nil
}

// revokeWarrant revokes payment warrant for specific service address
// It will be in effect next 1 hour, the will be deleted
// Params:
// - ParamServiceAddress address.Address
func revokeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := coreutil.NewAssert(ctx.Log())

	a.Require(ctx.Caller().IsAddress(), "payer must be an address")
	payerAddr := ctx.Caller().MustAddress()
	serviceAddr := par.MustGetAddress(ParamServiceAddress)

	w, r := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, a)
	a.Require(w > 0, fmt.Sprintf("warrant of %s to %s does not exist", payerAddr, serviceAddr))
	a.Require(r == 0, fmt.Sprintf("warrant of %s to %s is already being revoked", payerAddr, serviceAddr))

	revokeDeadline := getRevokeDeadline(ctx.GetTimestamp())
	setWarrantRevoke(ctx.State(), payerAddr, serviceAddr, revokeDeadline.Unix())

	succ := ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: ctx.ContractID(),
		EntryPoint:       coretypes.Hn(FuncCloseWarrant),
		TimeLock:         uint32(revokeDeadline.Unix()),
		Params: codec.MakeDict(map[string]interface{}{
			ParamPayerAddress:   payerAddr,
			ParamServiceAddress: serviceAddr,
		}),
	})
	a.Require(succ, "failed to post time-locked 'closeWarrant' request to self")
	return nil, nil
}

// closeWarrant can only be sent from self. It closes the warrant account
// - ParamServiceAddress address.Address
// - ParamPayerAddress address.Address
func closeWarrant(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := coreutil.NewAssert(ctx.Log())
	a.Require(ctx.Caller() == coretypes.NewAgentIDFromContractID(ctx.ContractID()), "caller must be self")

	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	warrant, _ := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, coreutil.NewAssert(ctx.Log()))
	if warrant > 0 {
		succ := ctx.TransferToAddress(payerAddr, cbalances.NewIotasOnly(warrant))
		a.Require(succ, "failed to send %d iotas to address %s", warrant, payerAddr)
	}
	deleteWarrant(ctx.State(), payerAddr, serviceAddr)
	return nil, nil
}

// getWarrantInfo return warrant info for given payer and services addresses
// Params:
// - ParamServiceAddress address.Address
// - ParamPayerAddress address.Address
// Output:
// - ParamWarrant int64 if == 0 no warrant
// - ParamRevoked int64 is exists, timestamp in Unix nanosec when warrant will be revoked
func getWarrantInfo(ctx coretypes.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	payerAddr := par.MustGetAddress(ParamPayerAddress)
	serviceAddr := par.MustGetAddress(ParamServiceAddress)
	warrant, revoke := getWarrantInfoIntern(ctx.State(), payerAddr, serviceAddr, coreutil.NewAssert(ctx.Log()))
	ret := dict.New()
	if warrant > 0 {
		ret.Set(ParamWarrant, codec.EncodeInt64(warrant))
	}
	if revoke > 0 {
		ret.Set(ParamRevoked, codec.EncodeInt64(revoke))
	}
	return ret, nil
}

//  utility

func getWarrantInfoIntern(state kv.KVStoreReader, payer, service address.Address, a coreutil.Assert) (int64, int64) {
	payerInfo := collections.NewMapReadOnly(state, string(payer[:]))
	warrantBin, err := payerInfo.GetAt(service[:])
	a.RequireNoError(err)
	warrant, exists, err := codec.DecodeInt64(warrantBin)
	a.RequireNoError(err)
	if !exists {
		warrant = 0
	}
	revokeBin, err := payerInfo.GetAt(getRevokeKey(service))
	revoke, exists, err := codec.DecodeInt64(revokeBin)
	if !exists {
		revoke = 0
	}
	return warrant, revoke
}

func setWarrant(state kv.KVStore, payer, service address.Address, value int64) {
	payerInfo := collections.NewMap(state, string(payer[:]))
	payerInfo.MustSetAt(service[:], codec.EncodeInt64(value))
}

func setWarrantRevoke(state kv.KVStore, payer, service address.Address, deadline int64) {
	payerInfo := collections.NewMap(state, string(payer[:]))
	payerInfo.MustSetAt(getRevokeKey(service), codec.EncodeInt64(deadline))
}

func deleteWarrant(state kv.KVStore, payer, service address.Address) {
	payerInfo := collections.NewMap(state, string(payer[:]))
	payerInfo.MustDelAt(service[:])
	payerInfo.MustDelAt(getRevokeKey(service))
}

func getPublicKey(state kv.KVStoreReader, addr address.Address, a coreutil.Assert) []byte {
	pkRegistry := collections.NewMapReadOnly(state, StateVarPublicKeys)
	ret, err := pkRegistry.GetAt(addr[:])
	a.RequireNoError(err)
	return ret
}

func getRevokeKey(service address.Address) []byte {
	return []byte(string(service[:]) + "-revoke")
}

func getRevokeDeadline(nowis int64) time.Time {
	return time.Unix(0, nowis).Add(WarrantRevokePeriod)
}
