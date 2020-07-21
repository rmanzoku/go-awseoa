package kmsutil

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	awseoa "github.com/rmanzoku/go-awseoa"
)

func TransactOptsFromAddress(svc *kms.KMS, addr common.Address) (*bind.TransactOpts, error) {
	keyID, err := KeyIDFromAddress(svc, addr)
	if err != nil {
		return nil, err
	}

	return awseoa.NewKMSTransactor(svc, keyID)
}

func KeyIDFromAddress(svc *kms.KMS, addr common.Address) (string, error) {
	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(in)
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
