package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rmanzoku/go-awseoa/v2"
	"github.com/rmanzoku/go-awseoa/v2/kmsutil"
)

var (
	rpc = os.Getenv("RPC")

	valueEther float64 = 0
	from               = ""
	to                 = ""
	gasGwei    float64 = 1
)

func handler(ctx context.Context) (err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	fmt.Printf("Sending %f Ether from %s to %s gas: %f gwei\n", valueEther, from, to, gasGwei)

	ethcli, err := ethclient.Dial(rpc)
	if err != nil {
		return err
	}
	chainID, err := ethcli.NetworkID(ctx)
	if err != nil {
		return
	}

	if !common.IsHexAddress(from) {
		return fmt.Errorf("invalid from address: %s", from)
	}
	svc := kms.NewFromConfig(cfg)
	keyId, err := kmsutil.KeyIDFromAddress(svc, common.HexToAddress(from))
	if err != nil {
		return err
	}

	topts, err := awseoa.NewKMSTransactor(svc, keyId, chainID)
	if err != nil {
		return err
	}
	topts.Context = context.TODO()
	topts.GasPrice, err = awseoa.GweiToWei(gasGwei)
	if err != nil {
		return err
	}

	amount, err := awseoa.EtherToWei(valueEther)
	if err != nil {
		return err
	}

	if !common.IsHexAddress(to) {
		return fmt.Errorf("invalid to address: %s", to)
	}
	toAddr := common.HexToAddress(to)

	tx, err := awseoa.SendEther(ethcli, topts, toAddr, amount)
	if err != nil {
		return
	}

	fmt.Println(tx.Hash().String())
	return
}

func main() {
	ctx := context.TODO()
	flag.Float64Var(&valueEther, "value", valueEther, "Sending value unit Ether")
	flag.StringVar(&from, "from", from, "From address")
	flag.StringVar(&to, "to", to, "To address")
	flag.Parse()

	if err := handler(ctx); err != nil {
		panic(err)
	}
}
