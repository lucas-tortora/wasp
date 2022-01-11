package tests

import (
	"fmt"
	"testing"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type advancedInccounterEnv struct {
	*chainEnv
}

func setupAdvancedInccounterTest(t *testing.T, clusterSize int, committee []int) *advancedInccounterEnv {
	quorum := uint16((2*len(committee))/3 + 1)

	clu := newCluster(t, clusterSize)

	addr, err := clu.RunDKG(committee, quorum)
	require.NoError(t, err)

	t.Logf("generated state address: %s", addr.Base58())

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), committee, quorum, addr)
	require.NoError(t, err)
	t.Logf("deployed chainID: %s", chain.ChainID.Base58())

	description := "testing with inccounter"
	progHash := inccounter.Contract.ProgramHash

	params := make(map[string]interface{})
	params[inccounter.VarCounter] = codec.EncodeInt64(0)
	_, err = chain.DeployContract(incCounterSCName, progHash.String(), description, params)
	require.NoError(t, err)

	e := &env{t: t, clu: clu}
	chEnv := &chainEnv{
		env:   e,
		chain: chain,
	}
	waitUntil(t, chEnv.contractIsDeployed(incCounterSCName), clu.Config.AllNodes(), 50*time.Second, "contract to be deployed")
	return &advancedInccounterEnv{
		chainEnv: chEnv,
	}
}

func (e *chainEnv) printBlocks(expected int) {
	recs, err := e.chain.GetAllBlockInfoRecordsReverse()
	require.NoError(e.t, err)

	sum := 0
	for _, rec := range recs {
		e.t.Logf("---- block #%d: total: %d, off-ledger: %d, success: %d", rec.BlockIndex, rec.TotalRequests, rec.NumOffLedgerRequests, rec.NumSuccessfulRequests)
		sum += int(rec.TotalRequests)
	}
	e.t.Logf("Total requests processed: %d", sum)
	require.EqualValues(e.t, expected, sum)
}

func TestAccessNodesOnLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Run("cluster=10, N=4, req=8", func(t *testing.T) {
		const numRequests = 8
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=10, N=4, req=100", func(t *testing.T) {
		const numRequests = 100
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=15, N=4, req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const numRequests = 1000
		const numValidatorNodes = 4
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})

	t.Run("cluster=15, N=6, req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOnLedger(t, numRequests, numValidatorNodes, clusterSize)
	})
}

func testAccessNodesOnLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int) {
	cmt := util.MakeRange(0, numValidatorNodes)
	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	for i := 0; i < numRequests; i++ {
		client := e.createNewClient()

		_, err := client.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), util.MakeRange(0, clusterSize), 40*time.Second)

	e.printBlocks(numRequests + 3)
}

func TestAccessNodesOffLedger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("cluster=6,N=4,req=8", func(t *testing.T) {
		const waitFor = 20 * time.Second
		const numRequests = 8
		const numValidatorNodes = 4
		const clusterSize = 6
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=10,N=4,req=50", func(t *testing.T) {
		const waitFor = 20 * time.Second
		const numRequests = 50
		const numValidatorNodes = 4
		const clusterSize = 10
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=10,N=6,req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 120 * time.Second
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 10
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=15,N=6,req=1000", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 120 * time.Second
		const numRequests = 1000
		const numValidatorNodes = 6
		const clusterSize = 15
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=30,N=15,req=8", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 60 * time.Second
		const numRequests = 8
		const numValidatorNodes = 15
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})

	t.Run("cluster=30,N=20,req=8", func(t *testing.T) {
		testutil.RunHeavy(t)
		const waitFor = 60 * time.Second
		const numRequests = 8
		const numValidatorNodes = 20
		const clusterSize = 30
		testAccessNodesOffLedger(t, numRequests, numValidatorNodes, clusterSize, waitFor)
	})
}

func testAccessNodesOffLedger(t *testing.T, numRequests, numValidatorNodes, clusterSize int, timeout ...time.Duration) {
	to := 60 * time.Second
	if len(timeout) > 0 {
		to = timeout[0]
	}
	cmt := util.MakeRange(0, numValidatorNodes)

	e := setupAdvancedInccounterTest(t, clusterSize, cmt)

	keyPair, myAddress := e.getOrCreateAddress()

	myAgentID := iscp.NewAgentID(myAddress, 0)

	accountsClient := e.chain.SCClient(accounts.Contract.Hname(), keyPair)
	_, err := accountsClient.PostRequest(accounts.FuncDeposit.Name, chainclient.PostRequestParams{
		Transfer: colored.NewBalancesForIotas(100),
	})
	require.NoError(t, err)

	waitUntil(t, e.balanceOnChainIotaEquals(myAgentID, 100), util.MakeRange(0, clusterSize), 60*time.Second, "send 100i")

	myClient := e.chain.SCClient(iscp.Hn(incCounterSCName), keyPair)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostOffLedgerRequest(inccounter.FuncIncCounter.Name, chainclient.PostRequestParams{Nonce: uint64(i + 1)})
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), util.MakeRange(0, clusterSize), to, "requests counted")

	e.printBlocks(numRequests + 4)
}

// extreme test
func TestAccessNodesMany(t *testing.T) {
	testutil.RunHeavy(t)
	const clusterSize = 15
	const numValidatorNodes = 6
	const requestsCountInitial = 2
	const requestsCountProgression = 2
	const iterationCount = 9

	e := setupAdvancedInccounterTest(t, clusterSize, util.MakeRange(0, numValidatorNodes))

	keyPair, _ := e.getOrCreateAddress()

	myClient := e.chain.SCClient(incCounterSCHname, keyPair)

	requestsCount := requestsCountInitial
	requestsCumulative := 0
	posted := 0
	for i := 0; i < iterationCount; i++ {
		logMsg := fmt.Sprintf("iteration %v of %v requests", i, requestsCount)
		t.Logf("Running %s", logMsg)
		for j := 0; j < requestsCount; j++ {
			_, err := myClient.PostRequest(inccounter.FuncIncCounter.Name)
			require.NoError(t, err)
		}
		posted += requestsCount
		requestsCumulative += requestsCount
		waitUntil(t, e.counterEquals(int64(requestsCumulative)), e.clu.Config.AllNodes(), 60*time.Second, logMsg)
		requestsCount *= requestsCountProgression
	}
	e.printBlocks(posted + 3)
}

// cluster of 10 access nodes and two overlapping committees
func TestRotation(t *testing.T) {
	numRequests := 8

	cmt1 := []int{0, 1, 2, 3}
	cmt2 := []int{2, 3, 4, 5}

	clu := newCluster(t, 10)
	addr1, err := clu.RunDKG(cmt1, 3)
	require.NoError(t, err)
	addr2, err := clu.RunDKG(cmt2, 3)
	require.NoError(t, err)

	t.Logf("addr1: %s", addr1.Base58())
	t.Logf("addr2: %s", addr2.Base58())

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt1, 3, addr1)
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID.Base58())

	description := "inccounter testing contract"
	programHash := inccounter.Contract.ProgramHash

	e := newChainEnv(t, clu, chain)

	_, err = chain.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, e.contractIsDeployed(incCounterSCName), e.clu.Config.AllNodes(), 30*time.Second)

	require.True(t, e.waitStateController(0, addr1, 5*time.Second))
	require.True(t, e.waitStateController(9, addr1, 5*time.Second))

	keyPair := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := ledgerstate.NewED25519Address(keyPair.PublicKey)
	e.requestFunds(myAddress, "myAddress")

	myClient := chain.SCClient(incCounterSCHname, &keyPair)

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(numRequests)), []int{0, 3, 8, 9}, 5*time.Second)

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair())

	params := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *params)

	require.NoError(t, err)

	require.True(t, e.waitBlockIndex(9, 4, 15*time.Second))
	require.True(t, e.waitBlockIndex(0, 4, 15*time.Second))
	require.True(t, e.waitBlockIndex(6, 4, 15*time.Second))

	reqid := iscp.NewRequestID(tx.ID(), 0)

	require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, 15*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, 15*time.Second))

	require.NoError(t, err)
	require.True(t, isAllowedStateControllerAddress(t, chain, 0, addr2))

	require.True(t, e.waitStateController(0, addr1, 15*time.Second))
	require.True(t, e.waitStateController(9, addr1, 15*time.Second))

	params = chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addr2).WithIotas(1)
	tx, err = govClient.PostRequest(governance.FuncRotateStateController.Name, *params)
	require.NoError(t, err)

	require.True(t, e.waitStateController(0, addr2, 15*time.Second))
	require.True(t, e.waitStateController(9, addr2, 15*time.Second))

	require.True(t, e.waitBlockIndex(9, 5, 15*time.Second))
	require.True(t, e.waitBlockIndex(0, 5, 15*time.Second))
	require.True(t, e.waitBlockIndex(6, 5, 15*time.Second))

	reqid = iscp.NewRequestID(tx.ID(), 0)
	require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, 15*time.Second))
	require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, 15*time.Second))

	for i := 0; i < numRequests; i++ {
		_, err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
		require.NoError(t, err)
	}

	waitUntil(t, e.counterEquals(int64(2*numRequests)), clu.Config.AllNodes(), 15*time.Second)
}

func TestRotationMany(t *testing.T) {
	testutil.RunHeavy(t)
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	const numRequests = 2
	const numCmt = 3
	const numRotations = 5
	const waitTimeout = 180 * time.Second

	cmtPredef := [][]int{
		{0, 1, 2, 3},
		{2, 3, 4, 5},
		{3, 4, 5, 6, 7, 8},
		{9, 4, 5, 6, 7, 8, 3},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}
	quorumPredef := []uint16{3, 3, 5, 5, 7}
	cmt := cmtPredef[:numCmt]
	quorum := quorumPredef[:numCmt]
	addrs := make([]iotago.Address, numCmt)

	var err error
	clu := newCluster(t, 10)
	for i := range cmt {
		addrs[i], err = clu.RunDKG(cmt[i], quorum[i])
		require.NoError(t, err)
		t.Logf("addr[%d]: %s", i, addrs[i].Base58())
	}

	chain, err := clu.DeployChain("chain", clu.Config.AllNodes(), cmt[0], quorum[0], addrs[0])
	require.NoError(t, err)
	t.Logf("chainID: %s", chain.ChainID.Base58())

	govClient := chain.SCClient(governance.Contract.Hname(), chain.OriginatorKeyPair())

	for i := range addrs {
		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addrs[i]).WithIotas(1)
		tx, err := govClient.PostRequest(governance.FuncAddAllowedStateControllerAddress.Name, *par)
		require.NoError(t, err)
		reqid := iscp.NewRequestID(tx.ID(), 0)
		require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain, 5, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, waitTimeout))
		require.True(t, isAllowedStateControllerAddress(t, chain, 0, addrs[i]))
		require.True(t, isAllowedStateControllerAddress(t, chain, 5, addrs[i]))
		require.True(t, isAllowedStateControllerAddress(t, chain, 9, addrs[i]))
	}

	description := "inccounter testing contract"
	programHash := inccounter.Contract.ProgramHash

	e := newChainEnv(t, clu, chain)

	_, err = chain.DeployContract(incCounterSCName, programHash.String(), description, nil)
	require.NoError(t, err)

	waitUntil(t, e.contractIsDeployed(incCounterSCName), clu.Config.AllNodes(), 30*time.Second)

	addrIndex := 0
	keyPair := cryptolib.NewKeyPairFromSeed(wallet.SubSeed(1))
	myAddress := ledgerstate.NewED25519Address(keyPair.PublicKey)
	e.requestFunds(myAddress, "myAddress")

	myClient := chain.SCClient(incCounterSCHname, &keyPair)

	for i := 0; i < numRotations; i++ {
		require.True(t, e.waitStateController(0, addrs[addrIndex], waitTimeout))
		require.True(t, e.waitStateController(4, addrs[addrIndex], waitTimeout))
		require.True(t, e.waitStateController(9, addrs[addrIndex], waitTimeout))

		for j := 0; j < numRequests; j++ {
			_, err = myClient.PostRequest(inccounter.FuncIncCounter.Name)
			require.NoError(t, err)
		}

		waitUntil(t, e.counterEquals(int64(numRequests*(i+1))), []int{0, 3, 8, 9}, 30*time.Second)

		addrIndex = (addrIndex + 1) % numCmt

		par := chainclient.NewPostRequestParams(governance.ParamStateControllerAddress, addrs[addrIndex]).WithIotas(1)
		tx, err := govClient.PostRequest(governance.FuncRotateStateController.Name, *par)
		require.NoError(t, err)
		reqid := iscp.NewRequestID(tx.ID(), 0)
		require.EqualValues(t, "", waitRequest(t, chain, 0, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain, 4, reqid, waitTimeout))
		require.EqualValues(t, "", waitRequest(t, chain, 9, reqid, waitTimeout))

		require.True(t, e.waitStateController(0, addrs[addrIndex], waitTimeout))
		require.True(t, e.waitStateController(4, addrs[addrIndex], waitTimeout))
		require.True(t, e.waitStateController(9, addrs[addrIndex], waitTimeout))
	}
}

func waitRequest(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid iscp.RequestID, timeout time.Duration) string {
	var ret string
	succ := waitTrue(timeout, func() bool {
		rec, err := callGetRequestRecord(t, chain, nodeIndex, reqid)
		if err == nil && rec != nil {
			ret = rec.ErrorStr
			return true
		}
		return false
	})
	if !succ {
		return "(timeout)"
	}
	return ret
}

func (e *chainEnv) waitBlockIndex(nodeIndex int, blockIndex uint32, timeout time.Duration) bool { //nolint:unparam // (timeout is always 5s)
	return waitTrue(timeout, func() bool {
		i, err := e.callGetBlockIndex(nodeIndex)
		return err == nil && i >= blockIndex
	})
}

func (e *chainEnv) callGetBlockIndex(nodeIndex int) (uint32, error) {
	ret, err := e.chain.Cluster.WaspClient(nodeIndex).CallView(
		e.chain.ChainID,
		blocklog.Contract.Hname(),
		blocklog.FuncGetLatestBlockInfo.Name,
		nil,
	)
	if err != nil {
		return 0, err
	}
	v, err := codec.DecodeUint32(ret.MustGet(blocklog.ParamBlockIndex))
	require.NoError(e.t, err)
	return v, nil
}

func callGetRequestRecord(t *testing.T, chain *cluster.Chain, nodeIndex int, reqid iscp.RequestID) (*blocklog.RequestReceipt, error) {
	args := dict.New()
	args.Set(blocklog.ParamRequestID, reqid.Bytes())

	res, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		blocklog.Contract.Hname(),
		blocklog.FuncGetRequestReceipt.Name,
		args,
	)
	if err != nil {
		return nil, xerrors.New("not found")
	}
	if len(res) == 0 {
		return nil, nil
	}
	recBin := res.MustGet(blocklog.ParamRequestRecord)
	rec, err := blocklog.RequestReceiptFromBytes(recBin)
	require.NoError(t, err)
	return rec, nil
}

func (e *chainEnv) waitStateController(nodeIndex int, addr iotago.Address, timeout time.Duration) bool {
	return waitTrue(timeout, func() bool {
		a, err := e.callGetStateController(nodeIndex)
		return err == nil && a.Equals(addr)
	})
}

func (e *chainEnv) callGetStateController(nodeIndex int) (iotago.Address, error) {
	ret, err := e.chain.Cluster.WaspClient(nodeIndex).CallView(
		e.chain.ChainID,
		blocklog.Contract.Hname(),
		blocklog.FuncControlAddresses.Name,
		nil,
	)
	if err != nil {
		return nil, err
	}
	addr, err := codec.DecodeAddress(ret.MustGet(blocklog.ParamStateControllerAddress))
	require.NoError(e.t, err)
	return addr, nil
}

func isAllowedStateControllerAddress(t *testing.T, chain *cluster.Chain, nodeIndex int, addr iotago.Address) bool {
	ret, err := chain.Cluster.WaspClient(nodeIndex).CallView(
		chain.ChainID,
		governance.Contract.Hname(),
		governance.FuncGetAllowedStateControllerAddresses.Name,
		nil,
	)
	require.NoError(t, err)
	arr := collections.NewArray16ReadOnly(ret, governance.ParamAllowedStateControllerAddresses)
	arrlen := arr.MustLen()
	if arrlen == 0 {
		return false
	}
	for i := uint16(0); i < arrlen; i++ {
		a, err := codec.DecodeAddress(arr.MustGetAt(i))
		require.NoError(t, err)
		if a.Equals(addr) {
			return true
		}
	}
	return false
}
