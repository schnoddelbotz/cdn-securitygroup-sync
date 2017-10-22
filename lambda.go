package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
)

// Handle is entrypoint for AWS lambda execution
func Handle(evt interface{}, ctx *runtime.Context) (string, error) {
	// disable timestamps for logging, as done by cloudwatch
	log.SetFlags(0)
	// get runtime configuration from parameter store
	ssmGet(os.Getenv("SSM_SOURCE"))
	// finally execute CLI-like
	run()
	return "Success", nil
}

func ssmGet(prefix string) {
	sess := session.Must(session.NewSession())
	sessConfig := &aws.Config{}
	svc := ssm.New(sess, sessConfig)
	params := &ssm.GetParametersInput{
		Names: []*string{
			aws.String(prefix + "_CSS_ARGS"),
			aws.String(prefix + "_AWS_SECGROUP_ID"),
			aws.String(prefix + "_AKAMAI_SSID"),
			aws.String(prefix + "_AKAMAI_EDGEGRID_HOST"),
			aws.String(prefix + "_AKAMAI_EDGEGRID_CLIENT_TOKEN"),
			aws.String(prefix + "_AKAMAI_EDGEGRID_CLIENT_SECRET"),
			aws.String(prefix + "_AKAMAI_EDGEGRID_ACCESS_TOKEN"),
		},
		WithDecryption: aws.Bool(true),
	}
	resp, err := svc.GetParameters(params)
	exitIfError("GetParameters", err)

	for _, p := range resp.Parameters {
		realName := (*p.Name)[len(prefix)+1:]
		switch realName {
		case "CSS_ARGS":
			parseLambdaFlags(*p.Value)
		case "AWS_SECGROUP_ID":
			mySecGroup = *p.Value
		case "AKAMAI_SSID":
			_ssid, _ := strconv.Atoi(*p.Value)
			mySSID = _ssid
		default:
			os.Setenv(realName, *p.Value)
		}
	}
}

func parseLambdaFlags(args string) {
	if strings.Contains(args, "-cloudflare") {
		useCloudflare = true
	}
	if strings.Contains(args, "-add-missing") {
		addMissing = true
	}
	if strings.Contains(args, "-delete-obsolete") {
		deleteObsolete = true
	}
	if strings.Contains(args, "-acknowledge") {
		acknowledge = true
	}
	if strings.Contains(args, "-list-ss-ids") {
		listSSIDs = true
	}
	if strings.Contains(args, "-version") {
		printVersion = true
	}
}
