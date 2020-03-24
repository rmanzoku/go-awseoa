package kmstransactor

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
	"github.com/ethereum/go-ethereum/core/types"
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

var topts *bind.TransactOpts

func TestFrom(t *testing.T) {
	initTesting(t)
	fmt.Println(topts.From.String())
}

func TestSendEther(t *testing.T) {
	is := initTesting(t)

	topts.GasPrice, _ = new(big.Int).SetString("1000000000", 10)
	topts.Context = context.TODO()

	amount, _ := new(big.Int).SetString("1000000000000", 10)

	ethcli, err := ethclient.Dial(rpc)
	is.Nil(err)

	tx, err := sendEther(ethcli, topts, to, amount)
	is.Nil(err)

	fmt.Println(tx.Hash().String())
}

func sendEther(client *ethclient.Client, transactOpts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	ctx := transactOpts.Context
	nonce, err := client.NonceAt(ctx, transactOpts.From, nil)
	if err != nil {
		return nil, err
	}
	tx := types.NewTransaction(nonce, to, amount, 21000, transactOpts.GasPrice, nil)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	tx, err = transactOpts.Signer(types.NewEIP155Signer(chainID), transactOpts.From, tx)
	if err != nil {
		return nil, err
	}

	return tx, client.SendTransaction(transactOpts.Context, tx)
}

func initTesting(t *testing.T) is.I {
	is := is.New(t)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(region)},
		Profile: profile,
	})

	is.Nil(err)
	k := kms.New(sess)

	topts, err = NewKMSTransactor(k, keyID)
	is.Nil(err)

	return is
}
