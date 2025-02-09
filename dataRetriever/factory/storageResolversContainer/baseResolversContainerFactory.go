package storageResolversContainers

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/storageResolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
)

const defaultBeforeGracefulClose = time.Minute

type baseResolversContainerFactory struct {
	container                  dataRetriever.ResolversContainer
	shardCoordinator           sharding.Coordinator
	messenger                  dataRetriever.TopicMessageHandler
	store                      dataRetriever.StorageService
	marshalizer                marshal.Marshalizer
	hasher                     hashing.Hasher
	uint64ByteSliceConverter   typeConverters.Uint64ByteSliceConverter
	dataPacker                 dataRetriever.DataPacker
	manualEpochStartNotifier   dataRetriever.ManualEpochStartNotifier
	chanGracefullyClose        chan endProcess.ArgEndProcess
	generalConfig              config.Config
	shardIDForTries            uint32
	chainID                    string
	workingDir                 string
	disableOldTrieStorageEpoch uint32
	epochNotifier              trie.EpochNotifier
}

func (brcf *baseResolversContainerFactory) checkParams() error {
	if check.IfNil(brcf.shardCoordinator) {
		return dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(brcf.messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(brcf.store) {
		return dataRetriever.ErrNilStore
	}
	if check.IfNil(brcf.marshalizer) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(brcf.uint64ByteSliceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(brcf.dataPacker) {
		return dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(brcf.manualEpochStartNotifier) {
		return dataRetriever.ErrNilManualEpochStartNotifier
	}
	if brcf.chanGracefullyClose == nil {
		return dataRetriever.ErrNilGracefullyCloseChannel
	}
	if check.IfNil(brcf.hasher) {
		return dataRetriever.ErrNilHasher
	}
	if check.IfNil(brcf.epochNotifier) {
		return dataRetriever.ErrNilEpochNotifier
	}

	return nil
}

func (brcf *baseResolversContainerFactory) generateTxResolvers(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards+1)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+1)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		resolver, err := brcf.createTxResolver(identifierTx, unit)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := brcf.createTxResolver(identifierTx, unit)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierTx

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createTxResolver(
	responseTopic string,
	unit dataRetriever.UnitType,
) (dataRetriever.Resolver, error) {

	txStorer := brcf.store.GetStorer(unit)

	arg := storageResolvers.ArgSliceResolver{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  txStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewSliceResolver(arg)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (brcf *baseResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := brcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards+2)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards+2)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
		resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierMiniBlocks
	}

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := brcf.createMiniBlocksResolver(identifierMiniBlocks)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards] = resolver
	keys[noOfShards] = identifierMiniBlocks

	identifierAllShardMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.AllShardId)
	allShardMiniblocksResolver, err := brcf.createMiniBlocksResolver(identifierAllShardMiniBlocks)
	if err != nil {
		return err
	}

	resolverSlice[noOfShards+1] = allShardMiniblocksResolver
	keys[noOfShards+1] = identifierAllShardMiniBlocks

	return brcf.container.AddMultiple(keys, resolverSlice)
}

func (brcf *baseResolversContainerFactory) createMiniBlocksResolver(responseTopic string) (dataRetriever.Resolver, error) {
	miniBlocksStorer := brcf.store.GetStorer(dataRetriever.MiniBlockUnit)

	arg := storageResolvers.ArgSliceResolver{
		Messenger:                brcf.messenger,
		ResponseTopicName:        responseTopic,
		Storage:                  miniBlocksStorer,
		DataPacker:               brcf.dataPacker,
		Marshalizer:              brcf.marshalizer,
		ManualEpochStartNotifier: brcf.manualEpochStartNotifier,
		ChanGracefullyClose:      brcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	mbResolver, err := storageResolvers.NewSliceResolver(arg)
	if err != nil {
		return nil, err
	}

	return mbResolver, nil
}

func (brcf *baseResolversContainerFactory) newImportDBTrieStorage(
	trieStorageConfig config.StorageConfig,
	mainStorer storage.Storer,
	checkpointsStorer storage.Storer,
) (common.StorageManager, dataRetriever.TrieDataGetter, error) {
	pathManager, err := storageFactory.CreatePathManager(
		storageFactory.ArgCreatePathManager{
			WorkingDir: brcf.workingDir,
			ChainID:    brcf.chainID,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		SnapshotDbCfg:            brcf.generalConfig.TrieSnapshotDB,
		Marshalizer:              brcf.marshalizer,
		Hasher:                   brcf.hasher,
		PathManager:              pathManager,
		TrieStorageManagerConfig: brcf.generalConfig.TrieStorageManagerConfig,
	}
	trieFactoryInstance, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	args := trieFactory.TrieCreateArgs{
		TrieStorageConfig:          trieStorageConfig,
		MainStorer:                 mainStorer,
		CheckpointsStorer:          checkpointsStorer,
		ShardID:                    core.GetShardIDString(brcf.shardIDForTries),
		PruningEnabled:             brcf.generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		CheckpointsEnabled:         brcf.generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:          brcf.generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
		DisableOldTrieStorageEpoch: brcf.disableOldTrieStorageEpoch,
		EpochStartNotifier:         brcf.epochNotifier,
	}
	return trieFactoryInstance.Create(args)
}
