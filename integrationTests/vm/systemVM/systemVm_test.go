package systemVM

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/stretchr/testify/assert"
)

func TestStakingUnstakingAndUnboundingOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,*:DEBUG")

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	verifyInitialBalance(t, nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	var txData string
	oneEncoded := hex.EncodeToString(big.NewInt(0).Bytes())
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "stake" + "@" + oneEncoded + "@" + pubKey + "@" + hex.EncodeToString([]byte("message"))
		integrationTests.CreateAndSendTransaction(node, node.EconomicsData.StakeValue(), factory.AuctionSCAddress, txData)
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	nonce, round = waitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	consumedBalance := big.NewInt(0).Add(big.NewInt(int64(len(txData))), big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit))
	consumedBalance.Mul(consumedBalance, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))

	checkAccountsAfterStaking(t, nodes, initialVal, consumedBalance)

	/////////------ send unStake tx
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unStake" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}
	consumed := big.NewInt(0).Add(big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit), big.NewInt(int64(len(txData))))
	consumed.Mul(consumed, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	consumedBalance.Add(consumedBalance, consumed)

	time.Sleep(time.Second)

	nonce, round = waitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	/////////----- wait for unbond period
	nonce, round = waitOperationToBeDone(t, nodes, int(nodes[0].EconomicsData.UnBondPeriod()), nonce, round, idxProposers)

	////////----- send unBond
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unBond" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}
	consumed = big.NewInt(0).Add(big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit), big.NewInt(int64(len(txData))))
	consumed.Mul(consumed, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	consumedBalance.Add(consumedBalance, consumed)

	time.Sleep(time.Second)

	_, _ = waitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	verifyUnbound(t, nodes, initialVal, consumedBalance)
}

func verifyUnbound(t *testing.T, nodes []*integrationTests.TestProcessorNode, initialVal, consumedBalance *big.Int) {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				expectedValue := big.NewInt(0).Sub(initialVal, consumedBalance)
				assert.Equal(t, expectedValue, sndAcc.Balance)
				break
			}
		}
	}
}

func checkAccountsAfterStaking(t *testing.T, nodes []*integrationTests.TestProcessorNode, initialVal, consumedBalance *big.Int) {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				expectedValue := big.NewInt(0).Sub(initialVal, node.EconomicsData.StakeValue())
				expectedValue = expectedValue.Sub(expectedValue, consumedBalance)
				assert.Equal(t, expectedValue, sndAcc.Balance)
				break
			}
		}
	}
}

func verifyInitialBalance(t *testing.T, nodes []*integrationTests.TestProcessorNode, initialVal *big.Int) {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				assert.Equal(t, initialVal, sndAcc.Balance)
				break
			}
		}
	}
}

func waitOperationToBeDone(t *testing.T, nodes []*integrationTests.TestProcessorNode, nrOfRounds int, nonce, round uint64, idxProposers []int) (uint64, uint64) {
	for i := 0; i < nrOfRounds; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	return nonce, round
}

func getAccountFromAddrBytes(accState state.AccountsAdapter, address []byte) *state.Account {
	addrCont, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(address)
	sndrAcc, _ := accState.GetExistingAccount(addrCont)

	sndAccSt, _ := sndrAcc.(*state.Account)

	return sndAccSt
}

func generateUniqueKey(identifier int) string {
	neededLength := 256
	uniqueIdentifier := fmt.Sprintf("%d", identifier)
	return strings.Repeat("0", neededLength-len(uniqueIdentifier)) + uniqueIdentifier
}
