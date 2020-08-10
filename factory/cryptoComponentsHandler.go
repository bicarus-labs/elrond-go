package factory

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ ComponentHandler = (*managedCryptoComponents)(nil)
var _ CryptoParamsHolder = (*managedCryptoComponents)(nil)
var _ CryptoComponentsHolder = (*managedCryptoComponents)(nil)
var _ CryptoComponentsHandler = (*managedCryptoComponents)(nil)

// CryptoComponentsHandlerArgs holds the arguments required to create a crypto components handler
type CryptoComponentsHandlerArgs CryptoComponentsFactoryArgs

// managedCryptoComponents creates the crypto components handler that can create, close and access the crypto components
type managedCryptoComponents struct {
	*cryptoComponents
	cryptoComponentsFactory *cryptoComponentsFactory
	mutCryptoComponents     sync.RWMutex
}

// NewManagedCryptoComponents creates a new Crypto components handler
func NewManagedCryptoComponents(args CryptoComponentsHandlerArgs) (*managedCryptoComponents, error) {
	ccf, err := NewCryptoComponentsFactory(CryptoComponentsFactoryArgs(args))
	if err != nil {
		return nil, err
	}

	return &managedCryptoComponents{
		cryptoComponents:        nil,
		cryptoComponentsFactory: ccf,
	}, nil
}

// Create creates the crypto components
func (mcc *managedCryptoComponents) Create() error {
	cc, err := mcc.cryptoComponentsFactory.Create()
	if err != nil {
		return err
	}

	mcc.mutCryptoComponents.Lock()
	mcc.cryptoComponents = cc
	mcc.mutCryptoComponents.Unlock()

	return nil
}

// Close closes the managed crypto components
func (mcc *managedCryptoComponents) Close() error {
	mcc.mutCryptoComponents.Lock()
	defer mcc.mutCryptoComponents.Unlock()

	if mcc.cryptoComponents != nil {
		err := mcc.cryptoComponents.Close()
		if err != nil {
			return err
		}
		mcc.cryptoComponents = nil
	}

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mcc *managedCryptoComponents) CheckSubcomponents() error {
	mcc.mutCryptoComponents.Lock()
	defer mcc.mutCryptoComponents.Unlock()

	if mcc.cryptoComponents == nil {
		return ErrNilCryptoComponents
	}
	if check.IfNil(mcc.cryptoComponents.publicKey) {
		return ErrNilPublicKey
	}
	if check.IfNil(mcc.cryptoComponents.privateKey) {
		return ErrNilPrivateKey
	}
	if check.IfNil(mcc.cryptoComponents.txSingleSigner) {
		return ErrNilTxSigner
	}
	if check.IfNil(mcc.cryptoComponents.blockSingleSigner) {
		return ErrNilBlockSigner
	}
	if check.IfNil(mcc.cryptoComponents.multiSigner) {
		return ErrNilMultiSigner
	}
	if check.IfNil(mcc.cryptoComponents.peerSignHandler) {
		return ErrNilPeerSignHandler
	}
	if check.IfNil(mcc.cryptoComponents.blockSignKeyGen) {
		return ErrNilBlockSignKeyGen
	}
	if check.IfNil(mcc.cryptoComponents.txSignKeyGen) {
		return ErrNilTxSignKeyGen
	}
	if check.IfNil(mcc.cryptoComponents.messageSignVerifier) {
		return ErrNilMessageSignVerifier
	}

	return nil
}

// PublicKey returns the configured validator public key
func (mcc *managedCryptoComponents) PublicKey() crypto.PublicKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.publicKey
}

// PrivateKey returns the configured validator private key
func (mcc *managedCryptoComponents) PrivateKey() crypto.PrivateKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.privateKey
}

// PublicKeyString returns the configured validator public key as string
func (mcc *managedCryptoComponents) PublicKeyString() string {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return ""
	}

	return mcc.cryptoParams.publicKeyString
}

// PublicKeyBytes returns the configured validator public key bytes
func (mcc *managedCryptoComponents) PublicKeyBytes() []byte {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.publicKeyBytes
}

// PrivateKeyBytes returns the configured validator private key bytes
func (mcc *managedCryptoComponents) PrivateKeyBytes() []byte {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.privateKeyBytes
}

// TxSingleSigner returns the transaction signer
func (mcc *managedCryptoComponents) TxSingleSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.txSingleSigner
}

// BlockSigner returns block single signer
func (mcc *managedCryptoComponents) BlockSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.blockSingleSigner
}

// MultiSigner returns the block multi-signer
func (mcc *managedCryptoComponents) MultiSigner() crypto.MultiSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.multiSigner
}

// PeerSignatureHandler returns the peer signature handler
func (mcc *managedCryptoComponents) PeerSignatureHandler() crypto.PeerSignatureHandler {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.peerSignHandler
}

// SetMultiSigner sets the block multi-signer
func (mcc *managedCryptoComponents) SetMultiSigner(ms crypto.MultiSigner) error {
	mcc.mutCryptoComponents.Lock()
	defer mcc.mutCryptoComponents.Unlock()

	if mcc.cryptoComponents == nil {
		return ErrNilCryptoComponents
	}

	mcc.cryptoComponents.multiSigner = ms
	return nil
}

// BlockSignKeyGen returns the block signer key generator
func (mcc *managedCryptoComponents) BlockSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.blockSignKeyGen
}

// TxSignKeyGen returns the transaction signer key generator
func (mcc *managedCryptoComponents) TxSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.txSignKeyGen
}

// MessageSignVerifier returns the message signature verifier
func (mcc *managedCryptoComponents) MessageSignVerifier() vm.MessageSignVerifier {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.messageSignVerifier
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcc *managedCryptoComponents) IsInterfaceNil() bool {
	return mcc == nil
}
