package arwenvm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestStorageForDistribution(t *testing.T) {
	// Only a test to benchmark storage on distribution contract
	//t.Skip()
	_ = logger.SetLogLevel("*:ERROR")
	ownerAddressBytes := []byte("12345678901234567890123456789011")
	ownerNonce := uint64(0)
	ownerBalance := big.NewInt(0).SetBytes([]byte("9999999999999999999999999999999999999999999999999999999999999999"))
	gasPrice := uint64(1)
	gasLimit := uint64(500000000)

	scCode := arwen.GetSCCode("../testdata/distribution.wasm")

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV4.toml")
	testContext, err := vm.CreateTxProcessorArwenVMWithGasSchedule(
		ownerNonce,
		ownerAddressBytes,
		ownerBalance,
		gasSchedule,
		vm.ArgEnableEpoch{},
	)
	require.Nil(t, err)
	defer testContext.Close()

	anotherAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, 100, factory.ArwenVirtualMachine)
	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	tx := vm.CreateDeployTx(
		ownerAddressBytes,
		ownerNonce,
		big.NewInt(0),
		gasPrice,
		gasLimit,
		arwen.CreateDeployTxData(scCode)+"@"+hex.EncodeToString([]byte("TOKEN-010101"))+"@"+hex.EncodeToString(anotherAddress),
	)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 10000000, []byte("startGlobalOperation"))
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 10000000, []byte("setCommunityDistribution@0781A4DA9A1E3D1C0007271E30FFFFFFFFFF@01ce"))
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, returnCode, vmcommon.Ok)
	ownerNonce++

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	numAddresses := 250000
	testAddresses := createTestAddresses(uint64(numAddresses))
	fmt.Println("SETUP DONE")

	valueToDistribute := big.NewInt(0).Mul(big.NewInt(10000000000000), big.NewInt(10000000000000))

	userPerStep := 100
	totalSteps := numAddresses / userPerStep
	fmt.Printf("Need to process %d transactions \n", totalSteps)

	for i := 0; i < numAddresses/userPerStep; i++ {
		start := time.Now()
		txData := "setPerUserDistributedLockedAssets@01ce"

		for j := i * userPerStep; j < (i+1)*userPerStep; j++ {
			txData += "@" + hex.EncodeToString(testAddresses[j]) + "@" + hex.EncodeToString(valueToDistribute.Bytes())
		}

		tx = vm.CreateTransaction(ownerNonce, big.NewInt(0), ownerAddressBytes, scAddress, gasPrice, 1000000000, []byte(txData))

		returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Equal(t, returnCode, vmcommon.Ok)
		ownerNonce++

		elapsedTime := time.Since(start)
		fmt.Printf("ID: %d, time elapsed to process 300 distribution %s \n", i, elapsedTime.String())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)
	}

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	accnt, err := testContext.Accounts.GetExistingAccount(scAddress)
	require.Nil(t, err)

	userAccnt := accnt.(vmcommon.UserAccountHandler)

	globalMap := smartContract.GlobalStorageMap
	keySize := uint64(0)
	valueSize := uint64(0)
	for key, val := range globalMap {
		keySize += uint64(len(key))
		valueSize += uint64(len(val))
	}

	fmt.Printf("KEY   SIZE %d \n", keySize)
	fmt.Printf("VALUE SIZE %d \n", valueSize)

	err = doTraceFile(userAccnt.GetRootHash(), testContext.Trie)
	require.Nil(t, err)
}

func doTraceFile(roothash []byte, tr common.Trie) error {
	log.Warn("saving trie trace file")

	traceFile, err := core.CreateFile(core.ArgCreateFileArgument{
		Directory:     "",
		Prefix:        "TRACE",
		FileExtension: "log",
	})
	if err != nil {
		return err
	}
	defer func() {
		_ = traceFile.Close()
	}()

	keysBytes := uint64(0)
	valueBytes := uint64(0)

	ch, err := tr.GetAllLeavesOnChannel(roothash)
	if err != nil {
		return err
	}
	for keyVal := range ch {
		_, err = traceFile.WriteString(fmt.Sprintf("%s : %s\n", hex.EncodeToString(keyVal.Key()), hex.EncodeToString(keyVal.Value())))
		if err != nil {
			return err
		}

		keysBytes += uint64(len(keyVal.Key()))
		valueBytes += uint64(len(keyVal.Value()))
	}

	_, err = traceFile.WriteString(fmt.Sprintf("TOTAL:\n  keys: %s\n  values: %s\n", core.ConvertBytes(keysBytes), core.ConvertBytes(valueBytes)))

	return nil
}
