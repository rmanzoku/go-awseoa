package awseoa

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

var (
	rpc     = os.Getenv("RPC")
	chainID = big.NewInt(4)
	to      = common.HexToAddress("0xd868711BD9a2C6F1548F5f4737f71DA67d821090")
	keyID   = os.Getenv("KEYID")
)

var svc *kms.Client
var topts *bind.TransactOpts

func TestFrom(t *testing.T) {
	fmt.Println(topts.From.String())
}

func TestCreateSigner(t *testing.T) {
	if os.Getenv("CREATE") == "" {
		t.Skip()
	}

	s, err := CreateSigner(svc, chainID)
	fmt.Println(err)
	assert.Nil(t, err)

	fmt.Println(s.Address().String())
}

func TestSetAlias(t *testing.T) {
	s, err := NewSigner(svc, keyID, chainID)
	if err != nil {
		t.Fatal(err)
	}

	err = s.SetAlias(s.Address().String())
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendEther(t *testing.T) {
	topts.GasPrice, _ = new(big.Int).SetString("1000000000", 10)
	topts.Context = context.TODO()

	amount, _ := new(big.Int).SetString("1000000000000", 10)

	ethcli, err := ethclient.Dial(rpc)
	assert.Nil(t, err)

	tx, err := SendEther(ethcli, topts, to, amount)
	assert.Nil(t, err)

	fmt.Println(tx.Hash().String())
}

func TestEthereumSign(t *testing.T) {
	s, err := NewSigner(svc, keyID, chainID)
	assert.Nil(t, err)

	msg := "0xd75be5d1b23bc1c3c22c0708a5c822f927f1eb8d609d684ef91996fd2bf2bbda"
	msgb, err := decodeHex(msg)
	assert.Nil(t, err)

	hash := toEthSignedMessageHash(msgb)

	sig, err := s.EthereumSign(msgb)
	assert.Nil(t, err)

	fmt.Println(s.Address().String())
	fmt.Println(encodeToHex(sig))

	addr, err := recover(hash, sig)
	assert.Nil(t, err)
	fmt.Println(addr.String())
}

func TestMain(m *testing.M) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	svc = kms.NewFromConfig(cfg)
	topts, err = NewKMSTransactor(svc, keyID, chainID)
	if err != nil {
		panic(err)
	}

	status := m.Run()
	os.Exit(status)
}
