package awseoa

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
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

func SendEther(client *ethclient.Client, opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
		opts.Context = ctx
	}

	chainId, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = client.NonceAt(ctx, opts.From, nil)
		if err != nil {
			return nil, err
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}

	opts.Value = amount

	// Estimate GasLimit
	gasLimit := opts.GasLimit
	if opts.GasLimit == 0 {
		gasLimit = 21000
	}

	var rawTx *types.Transaction

	if opts.GasPrice != nil {
		rawTx = types.NewTransaction(nonce, to, amount, gasLimit, opts.GasPrice, nil)
	} else {
		// Only query for basefee if gasPrice not specified
		if head, errHead := client.HeaderByNumber(ctx, nil); errHead != nil {
			return nil, errHead
		} else if head.BaseFee != nil {

			// Estimate TipCap
			gasTipCap := opts.GasTipCap
			if gasTipCap == nil {
				tip, err := client.SuggestGasTipCap(ctx)
				if err != nil {
					return nil, err
				}
				gasTipCap = tip
			}

			// Estimate FeeCap
			gasFeeCap := opts.GasFeeCap
			if gasFeeCap == nil {

				gasFeeCap = new(big.Int).Add(
					gasTipCap,
					new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
				)
			}
			if gasFeeCap.Cmp(gasTipCap) < 0 {
				return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
			}

			baseTx := &types.DynamicFeeTx{
				ChainID:    chainId,
				Nonce:      nonce,
				GasTipCap:  gasTipCap,
				GasFeeCap:  gasFeeCap,
				Gas:        gasLimit,
				To:         &to,
				Value:      amount,
				Data:       nil,
				AccessList: nil,
			}

			rawTx = types.NewTx(baseTx)
		} else {
			gasPrice, err := client.SuggestGasPrice(ctx)
			if err != nil {
				return nil, err
			}
			// Chain is not London ready -> use legacy transaction
			rawTx = types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
		}
	}

	tx, err := opts.Signer(opts.From, rawTx)
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
