package kmsutil

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/common"
)

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
