package kmstransactor

import (
	"encoding/asn1"
	"errors"
	"math/big"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

func NewKMSTransactor(svc *kms.KMS, id string) (*bind.TransactOpts, error) {
	k := &key{
		KMS: svc,
		id:  id,
	}
	pub, err := k.Pubkey()
	if err != nil {
		return nil, err
	}

	keyAddr, err := publicKeyBytesToAddress(pub)
	if err != nil {
		return nil, err
	}

	return &bind.TransactOpts{
		From: keyAddr,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, errors.New("not authorized to sign this account")
			}

			digest := signer.Hash(tx).Bytes()

			sig, err := k.Sign(digest)
			if err != nil {
				return nil, err
			}

			// calc v
			for _, v := range []int{0, 1} {
				sigv := append(sig, byte(v))
				pubkey, err := secp256k1.RecoverPubkey(digest, sigv)
				if err != nil {
					return nil, err
				}

				candidate, err := publicKeyBytesToAddress(pubkey)
				if err != nil {
					return nil, err
				}

				if reflect.DeepEqual(keyAddr.Bytes(), candidate.Bytes()) {
					sig = append(sig, byte(v))
					break
				}
			}

			return tx.WithSignature(signer, sig)
		}}, nil
}

type key struct {
	*kms.KMS
	id     string
	pubkey []byte
}

func (k key) Pubkey() ([]byte, error) {
	if k.pubkey != nil {
		return k.pubkey, nil
	}
	in := &kms.GetPublicKeyInput{
		KeyId: aws.String(k.id),
	}
	out, err := k.GetPublicKey(in)
	if err != nil {
		return nil, err
	}

	seq := new(struct {
		Identifiers struct {
			KeyType asn1.ObjectIdentifier
			Curve   asn1.ObjectIdentifier
		}
		Pubkey asn1.BitString
	})
	_, err = asn1.Unmarshal(out.PublicKey, seq)
	if err != nil {
		return nil, err
	}

	return seq.Pubkey.Bytes, nil
}

func (k key) Sign(digest []byte) (signature []byte, err error) {
	in := &kms.SignInput{
		KeyId:            aws.String(k.id),
		Message:          digest,
		SigningAlgorithm: aws.String("ECDSA_SHA_256"),
		MessageType:      aws.String("DIGEST"),
	}
	out, err := k.KMS.Sign(in)
	if err != nil {
		return nil, err
	}

	sig := new(struct {
		R *big.Int
		S *big.Int
	})
	_, err = asn1.Unmarshal(out.Signature, sig)
	if err != nil {
		return nil, err
	}

	// EIP-2
	if sig.S.Cmp(secp256k1halfN) > 0 {
		sig.S = new(big.Int).Sub(secp256k1N, sig.S)
	}

	return append(sig.R.Bytes(), sig.S.Bytes()...), nil
}

func publicKeyBytesToAddress(pub []byte) (common.Address, error) {
	pubkey, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}
