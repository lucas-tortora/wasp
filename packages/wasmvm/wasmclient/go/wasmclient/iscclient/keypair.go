package iscclient

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"

	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type Keypair struct {
	privateKey []byte
	publicKey  []byte
}

func MakeSubSeed(seed []byte, index uint32) []byte {
	zero := []byte{0x00}
	buf := make([]byte, 4)
	hash := make([]byte, 0, 64)

	h := hmac.New(sha512.New, []byte("ed25519 seed"))
	h.Write(seed)
	hash = h.Sum(hash[:0])
	key := hash[:32]
	chainCode := hash[32:]

	coinType := uint32(1) // testnet
	switch HrpForClient {
	case "iota":
		coinType = 4218
	case "smr":
		coinType = 4219
	}

	path := []uint32{44, coinType, index, 0, 0}
	for _, element := range path {
		binary.BigEndian.PutUint32(buf, element|0x80000000)
		h = hmac.New(sha512.New, chainCode)
		h.Write(zero)
		h.Write(key)
		h.Write(buf)
		hash = h.Sum(hash[:0])
		key = hash[:32]
		chainCode = hash[32:]
	}
	return key
}

func KeyPairFromSeed(seed []byte) *Keypair {
	pair := ed25519.NewKeyFromSeed(seed)
	return &Keypair{
		privateKey: pair,
		publicKey:  pair[32:],
	}
}

func KeyPairFromSubSeed(seed []byte, index uint32) *Keypair {
	sub := MakeSubSeed(seed, index)
	return KeyPairFromSeed(sub)
}

func (kp *Keypair) Address() wasmtypes.ScAddress {
	address := blake2b.Sum256(kp.publicKey)
	buf := append([]byte{wasmtypes.ScAddressEd25519}, address[:]...)
	return wasmtypes.AddressFromBytes(buf)
}

func (kp *Keypair) GetPrivateKey() []byte {
	return kp.privateKey
}

func (kp *Keypair) GetPublicKey() []byte {
	return kp.publicKey
}
