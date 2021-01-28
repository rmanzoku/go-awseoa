# AWS Managed EOA
AWS Managed EOA is an Ethereum EOA(Externally Owned Account) using [Asymmetric Keys of AWS Key Management Service](https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html).

Now `aws-sdk-go-v2` is supported!. you can use v2

## Using commad line

```sh
$ go install github.com/rmanzoku/go-awseoa/cmd/awseoa

$ export AWS_REGION=YOUR_REGION
$ export AWS_PROFILE=YOUR_PROFILE
$ awseoa list
# list keys

$ awseoa new
# create new key and set alias as address
```

## Example for sending Ether v2

```golang
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rmanzoku/go-awseoa/v2"
)

var (
	rpc   = os.Getenv("RPC")
	to    = common.HexToAddress("0xd868711BD9a2C6F1548F5f4737f71DA67d821090")
	keyID = os.Getenv("KEYID")
)

func main() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	svc := kms.NewFromConfig(cfg)

	ethcli, _ := ethclient.Dial(rpc)
	chainID, err := ethcli.NetworkID(ctx)
	if err != nil {
		panic(err)
	}

	topts, _ := awseoa.NewKMSTransactor(svc, keyID, chainID)
	topts.GasPrice, _ = new(big.Int).SetString("1000000000", 10)
	topts.Context = context.TODO()

	amount, _ := new(big.Int).SetString("1000000000000", 10)

	tx, err := sendEther(ethcli, topts, to, amount)
	if err != nil {
		panic(err)
	}

	fmt.Println(tx.Hash().String())
}

func sendEther(client *ethclient.Client, transactOpts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	ctx := transactOpts.Context
	nonce, err := client.NonceAt(ctx, transactOpts.From, nil)
	if err != nil {
		return nil, err
	}
	tx := types.NewTransaction(nonce, to, amount, 21000, transactOpts.GasPrice, nil)

	tx, err = transactOpts.Signer(transactOpts.From, tx)
	if err != nil {
		return nil, err
	}

	return tx, client.SendTransaction(transactOpts.Context, tx)
}
```

## Same concept
- A JavaScript Web3 Provider using AWS KMS [odanado/aws-kms-provider](https://github.com/odanado/aws-kms-provider)
