# AWS Managed EOA
AWS Managed EOA is an Ethereum EOA(Externally Owned Account) using [Asymmetric Keys of AWS Key Management Service](https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html).

## Example for sending Ether

```golang
package main

import (
        "os"

        "github.com/ethereum/go-ethereum/common"
)

var (
        rpc     = os.Getenv("RPC")
        profile = os.Getenv("AWS_PROFILE")
        region  = os.Getenv("AWS_DEFAULT_REGION")
        to      = common.HexToAddress("0xd868711BD9a2C6F1548F5f4737f71DA67d821090")
        keyID   = os.Getenv("KEYID")
)

func main() {
        sess, _ := session.NewSessionWithOptions(session.Options{
                Config:  aws.Config{Region: aws.String(region)},
                Profile: profile,
        })

        svc = kms.New(sess)

        topts, _ = NewKMSTransactor(svc, keyID)
        topts.GasPrice, _ = new(big.Int).SetString("1000000000", 10)
        topts.Context = context.TODO()

        amount, _ := new(big.Int).SetString("1000000000000", 10)

        ethcli, _ := ethclient.Dial(rpc)

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
```

## See also
- A JavaScript Web3 Provider using AWS KMS [odanado/aws-kms-provider](https://github.com/odanado/aws-kms-provider)
