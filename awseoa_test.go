package awseoa

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cheekybits/is"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/net/context"
)

var (
	rpc     = os.Getenv("RPC")
	profile = os.Getenv("AWS_PROFILE")
	region  = os.Getenv("AWS_DEFAULT_REGION")
	to      = common.HexToAddress("0xd868711BD9a2C6F1548F5f4737f71DA67d821090")
	keyID   = os.Getenv("KEYID")
)

var svc *kms.KMS
var topts *bind.TransactOpts

func TestFrom(t *testing.T) {
	initTesting(t)
	fmt.Println(topts.From.String())
}

func TestCreateSigner(t *testing.T) {
	if os.Getenv("CREATE") == "" {
		t.Skip()
	}
	is := initTesting(t)
	s, err := CreateSigner(svc)
	fmt.Println(err)
	is.Nil(err)

	fmt.Println(s.Address().String())
}

func TestSetAlias(t *testing.T) {
	is := initTesting(t)
	s, err := NewSigner(svc, keyID)
	is.Nil(err)

	err = s.SetAlias(s.Address().String())
	is.Nil(err)
}

func TestSendEther(t *testing.T) {
	is := initTesting(t)

	topts.GasPrice, _ = new(big.Int).SetString("1000000000", 10)
	topts.Context = context.TODO()

	amount, _ := new(big.Int).SetString("1000000000000", 10)

	ethcli, err := ethclient.Dial(rpc)
	is.Nil(err)

	tx, err := SendEther(ethcli, topts, to, amount)
	is.Nil(err)

	fmt.Println(tx.Hash().String())
}

func TestEthereumSign(t *testing.T) {
	is := initTesting(t)

	s, err := NewSigner(svc, keyID)
	is.Nil(err)

	msg := "0xd75be5d1b23bc1c3c22c0708a5c822f927f1eb8d609d684ef91996fd2bf2bbda"
	msgb, err := decodeHex(msg)
	is.Nil(err)

	hash := toEthSignedMessageHash(msgb)

	sig, err := s.EthereumSign(msgb)
	is.Nil(err)

	fmt.Println(s.Address().String())
	fmt.Println(encodeToHex(sig))

	addr, err := recover(hash, sig)
	fmt.Println(addr.String())
}

func initTesting(t *testing.T) is.I {
	is := is.New(t)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(region)},
		Profile: profile,
	})

	is.Nil(err)
	svc = kms.New(sess)

	topts, err = NewKMSTransactor(svc, keyID)
	is.Nil(err)

	return is
}
