package awseoa

import (
	"encoding/asn1"
	"errors"
	"fmt"
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

func NewKMSTransactor(svc *kms.KMS, id string, chainID *big.Int) (*bind.TransactOpts, error) {
	s, err := NewSigner(svc, id, chainID)
	if err != nil {
		return nil, err
	}

	pub, err := s.Pubkey()
	if err != nil {
		return nil, err
	}

	keyAddr, err := publicKeyBytesToAddress(pub)
	if err != nil {
		return nil, err
	}

	return &bind.TransactOpts{
		From: keyAddr,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, errors.New("not authorized to sign this account")
			}
			signer := types.NewEIP155Signer(s.chainID)
			digest := signer.Hash(tx).Bytes()

			sig, err := s.SignDigest(digest)
			if err != nil {
				return nil, err
			}

			return tx.WithSignature(signer, sig)
		}}, nil
}

type Signer struct {
	*kms.KMS
	ID      string
	pubkey  []byte
	chainID *big.Int
}

func NewSigner(svc *kms.KMS, id string, chainID *big.Int) (*Signer, error) {
	s := &Signer{KMS: svc, ID: id, pubkey: nil, chainID: chainID}
	_, err := s.Pubkey()
	return s, err
}

func CreateSigner(svc *kms.KMS, chainID *big.Int) (*Signer, error) {
	in := new(kms.CreateKeyInput)
	in.SetCustomerMasterKeySpec("ECC_SECG_P256K1")
	in.SetKeyUsage("SIGN_VERIFY")
	in.SetOrigin("AWS_KMS")

	out, err := svc.CreateKey(in)
	if err != nil {
		return nil, err
	}
	id := *out.KeyMetadata.KeyId
	s, err := NewSigner(svc, id, chainID)
	if err != nil {
		return nil, err
	}

	err = s.SetAlias(s.Address().String())
	return s, err
}

func (s Signer) Address() common.Address {
	pub, err := s.Pubkey()
	if err != nil {
		panic(err)
	}
	ret, err := publicKeyBytesToAddress(pub)
	if err != nil {
		panic(err)
	}

	return ret
}

func (s Signer) SetAlias(alias string) error {
	in := new(kms.CreateAliasInput)
	in.SetAliasName("alias/" + alias)
	in.SetTargetKeyId(s.ID)
	_, err := s.KMS.CreateAlias(in)
	return err
}

func (s Signer) Pubkey() ([]byte, error) {
	if s.pubkey != nil {
		return s.pubkey, nil
	}
	in := &kms.GetPublicKeyInput{
		KeyId: aws.String(s.ID),
	}
	out, err := s.KMS.GetPublicKey(in)
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

func (s Signer) SignDigest(digest []byte) (signature []byte, err error) {
	in := &kms.SignInput{
		KeyId:            aws.String(s.ID),
		Message:          digest,
		SigningAlgorithm: aws.String("ECDSA_SHA_256"),
		MessageType:      aws.String("DIGEST"),
	}
	out, err := s.KMS.Sign(in)
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

	signature = append(sig.R.Bytes(), sig.S.Bytes()...)

	// Calc V
	for _, v := range []int{0, 1} {
		sigv := append(signature, byte(v))
		pubkey, err := secp256k1.RecoverPubkey(digest, sigv)
		if err != nil {
			return nil, err
		}

		candidate, err := publicKeyBytesToAddress(pubkey)
		if err != nil {
			return nil, err
		}

		if reflect.DeepEqual(s.Address().Bytes(), candidate.Bytes()) {
			signature = append(signature, byte(v))
			break
		}
	}

	return signature, nil
}

func (s Signer) EthereumSign(msg []byte) (signature []byte, err error) {
	digest := toEthSignedMessageHash(msg)
	sig, err := s.SignDigest(digest)

	if sig[64] < 27 {
		sig[64] += 27
	}

	return sig, nil
}

func (s Signer) TransactOpts() (*bind.TransactOpts, error) {
	return NewKMSTransactor(s.KMS, s.ID, s.chainID)
}

func publicKeyBytesToAddress(pub []byte) (common.Address, error) {
	pubkey, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}

func toEthSignedMessageHash(message []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	return keccak256([]byte(msg))
}

func keccak256(data []byte) []byte {
	return crypto.Keccak256(data)
}
