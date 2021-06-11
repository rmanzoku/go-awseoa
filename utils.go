package awseoa

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/shopspring/decimal"
)

func recover(hash, sig []byte) (common.Address, error) {
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.HexToAddress("0x0"), err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

func encodeToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func decodeHex(s string) ([]byte, error) {
	if s[0:2] != "0x" {
		return nil, errors.New("hex must start with 0x")
	}
	return hex.DecodeString(s[2:])
}

func SendEther(client *ethclient.Client, transactOpts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
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

	return tx, client.SendTransaction(ctx, tx)
}

func EtherToWei(ether float64) (*big.Int, error) {
	etherDecimal := decimal.NewFromFloat(ether)
	base := decimal.NewFromInt(params.Ether)

	retDecimal := etherDecimal.Mul(base).Floor()
	ret, ok := new(big.Int).SetString(retDecimal.String(), 10)
	if !ok {
		return nil, errors.New("Invalit number " + retDecimal.String())
	}
	return ret, nil
}

func GweiToWei(gwei float64) (*big.Int, error) {
	gweiDecimal := decimal.NewFromFloat(gwei)
	base := decimal.NewFromInt(params.GWei)

	retDecimal := gweiDecimal.Mul(base).Floor()
	ret, ok := new(big.Int).SetString(retDecimal.String(), 10)
	if !ok {
		return nil, errors.New("Invalit number " + retDecimal.String())
	}
	return ret, nil
}
