package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/types"
)

// ScheduledDataSyncerFactory is a factory for the scheduled data syncer
type ScheduledDataSyncerFactory struct {
	config config.Config
}

// NewScheduledDataSyncerFactory creates a factory instance
func NewScheduledDataSyncerFactory(config config.Config) *ScheduledDataSyncerFactory {
	return &ScheduledDataSyncerFactory{config: config}
}

// Create creates a scheduled data syncer
func (sdsf *ScheduledDataSyncerFactory) Create(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
	return newStartInEpochShardHeaderDataSyncerWithScheduled(
		args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch, sdsf.config)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (sdsf *ScheduledDataSyncerFactory) IsInterfaceNil() bool {
	return sdsf == nil
}
