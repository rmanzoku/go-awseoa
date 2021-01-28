package kmsutil

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	awseoa "github.com/rmanzoku/go-awseoa"
)

func NewKMSClient() (*kms.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return kms.NewFromConfig(cfg), nil
}

func TransactOptsFromAddress(svc *kms.Client, addr common.Address, chainID *big.Int) (*bind.TransactOpts, error) {
	keyID, err := KeyIDFromAddress(svc, addr)
	if err != nil {
		return nil, err
	}

	return awseoa.NewKMSTransactor(svc, keyID, chainID)
}

func KeyIDFromAddress(svc *kms.Client, addr common.Address) (string, error) {
	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(context.TODO(), in)
	if err != nil {
		return "", err
	}

	for _, a := range out.Aliases {
		alias := "None"
		if a.AliasName != nil {
			alias = *a.AliasName
		}
		alias = strings.TrimPrefix(alias, "alias/")
		if strings.HasPrefix(alias, "aws/") {
			continue
		}

		ad := common.HexToAddress(alias)
		if ad.String() != addr.String() {
			continue
		}

		return *a.TargetKeyId, nil
	}

	return "", errors.New("Not found addr: " + addr.String())
}
