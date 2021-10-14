package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
)

func createInterceptedHeaderArg(header data.HeaderHandler) *interceptedBlocks.ArgInterceptedBlockHeader {
	args := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        &mock.ShardCoordinatorStub{},
		Hasher:                  &mock.HasherMock{},
		Marshalizer:             &mock.MarshalizerMock{},
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger:       &mock.EpochStartTriggerStub{},
	}

	args.HdrBuff, _ = args.Marshalizer.Marshal(header)

	return args
}

func CreateInterceptedHeaderData(header data.HeaderHandler) *interceptedBlocks.InterceptedHeader {
	args := createInterceptedHeaderArg(header)
	interceptedHeader, _ := interceptedBlocks.NewInterceptedHeader(args)

	return interceptedHeader
}
