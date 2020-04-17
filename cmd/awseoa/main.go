package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	awseoa "github.com/rmanzoku/go-awseoa"
)

var (
	region  = os.Getenv("AWS_REGION")
	profile = os.Getenv("AWS_PROFILE")
)

func List(svc *kms.KMS) (err error) {

	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(in)
	if err != nil {
		return
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
		keyID := "None"
		if a.TargetKeyId != nil {
			keyID = *a.TargetKeyId
		}

		fmt.Println(alias, keyID)
	}
	return
}

func New(svc *kms.KMS) (err error) {
	signer, err := awseoa.CreateSigner(svc)
	if err != nil {
		return
	}
	fmt.Println(signer.Address().String(), signer.ID)
	return
}

func main() {
	var err error
	_ = flag.NewFlagSet("list", flag.ExitOnError)
	_ = flag.NewFlagSet("new", flag.ExitOnError)

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(region)},
		Profile: profile,
	})
	if err != nil {
		return
	}
	svc := kms.New(sess)

	switch os.Args[1] {
	case "list":
		err = List(svc)
	case "new":
		err = New(svc)
	default:
		flag.Usage()
	}

	if err != nil {
		panic(err)
	}
}
