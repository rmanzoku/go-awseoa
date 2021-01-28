package main

import (
	"flag"
	"fmt"
	"math/big"
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

var (
	flagTags = false
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

		tags := ""
		if flagTags {
			in := &kms.ListResourceTagsInput{KeyId: a.TargetKeyId}
			out, err := svc.ListResourceTags(in)
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

func AddTag(svc *kms.KMS, keyID, tagKey, tagValue string) (err error) {
	in := &kms.TagResourceInput{
		KeyId: aws.String(keyID),
		Tags: []*kms.Tag{
			{
				TagKey:   aws.String(tagKey),
				TagValue: aws.String(tagValue),
			},
		},
	}
	_, err = svc.TagResource(in)
	return
}

func New(svc *kms.KMS) (err error) {
	signer, err := awseoa.CreateSigner(svc, big.NewInt(1))
	if err != nil {
		return
	}
	fmt.Println(signer.Address().String(), signer.ID)
	return
}

func main() {
	var err error
	listFlag := flag.NewFlagSet("list", flag.ExitOnError)
	_ = flag.NewFlagSet("new", flag.ExitOnError)
	_ = flag.NewFlagSet("add-tag", flag.ExitOnError)

	listFlag.BoolVar(&flagTags, "tags", flagTags, "Show tags")

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
		flag.Usage()
	}

	if err != nil {
		panic(err)
	}
}
