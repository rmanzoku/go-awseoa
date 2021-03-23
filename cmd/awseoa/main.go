package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	awseoa "github.com/rmanzoku/go-awseoa/v2"
	"github.com/rmanzoku/go-awseoa/v2/kmsutil"
)

var (
	flagTags = true
)

func List(svc *kms.Client) (err error) {

	in := &kms.ListAliasesInput{}
	out, err := svc.ListAliases(context.TODO(), in)
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

		tags := ""
		if flagTags {
			in := &kms.ListResourceTagsInput{KeyId: a.TargetKeyId}
			out, err := svc.ListResourceTags(context.TODO(), in)
			if err != nil {
				return err
			}

			for _, t := range out.Tags {
				tags += *t.TagKey + ":" + *t.TagValue + "\t"
			}
		}

		fmt.Println(alias, keyID, tags)
	}
	return
}

func AddTag(svc *kms.Client, keyID, tagKey, tagValue string) (err error) {
	in := &kms.TagResourceInput{
		KeyId: aws.String(keyID),
		Tags: []kmstypes.Tag{
			{
				TagKey:   aws.String(tagKey),
				TagValue: aws.String(tagValue),
			},
		},
	}
	_, err = svc.TagResource(context.TODO(), in)
	return
}

func New(svc *kms.Client) (err error) {
	signer, err := awseoa.CreateSigner(svc, big.NewInt(4))
	if err != nil {
		return
	}
	fmt.Println(signer.Address().String(), signer.ID)
	return
}

func usage() {
	fmt.Println("Usage of awseoa:")
	fmt.Println("")
	fmt.Println("   list     Show list of keys")
	fmt.Println("            --tags: with tags")
	fmt.Println("   new      Create key")
	fmt.Println("   add-tag [keyID] [name:value]")
	fmt.Println("            add tag to exist key")
}

func main() {
	var err error
	listFlag := flag.NewFlagSet("list", flag.ExitOnError)
	_ = flag.NewFlagSet("new", flag.ExitOnError)
	_ = flag.NewFlagSet("add-tag", flag.ExitOnError)

	listFlag.BoolVar(&flagTags, "tags", flagTags, "Show tags")

	if len(os.Args) == 1 {
		usage()
		return
	}

	svc, err := kmsutil.NewKMSClient()
	if err != nil {
		panic(err)
	}
	listFlag.Parse(os.Args[2:])

	switch os.Args[1] {
	case "list":
		err = List(svc)
	case "new":
		err = New(svc)
	case "add-tag":
		keyID := os.Args[2]
		tag := strings.Split(os.Args[3], ":")
		err = AddTag(svc, keyID, tag[0], tag[1])
	default:
		usage()
	}

	if err != nil {
		panic(err)
	}
}
