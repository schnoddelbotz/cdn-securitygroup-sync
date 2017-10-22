package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getSecGroupCIDRs(secGroupID string) map[string]struct{} {
	secGroups := []string{secGroupID}
	cidrs := make(map[string]struct{})
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)

	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice(secGroups),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidGroupId.Malformed":
				exitErrorf("%s.", aerr.Message())
			case "InvalidGroup.NotFound":
				exitErrorf("%s. Maybe wrong AWS_REGION ?", aerr.Message())
			}
		}
		exitErrorf("Unable to get descriptions for security groups, %v", err)
	}

	// only care about rules with port range of 80-443
	for _, group := range result.SecurityGroups {
		for _, perm := range group.IpPermissions {
			if *perm.FromPort == 80 && *perm.ToPort == 443 {
				for _, addr := range perm.IpRanges {
					cidrs[*addr.CidrIp] = struct{}{}
				}
			}
		}
	}

	return cidrs
}

func addMissingCIDRs(secGroupID string, CIDRs []string) {
	if len(CIDRs) == 0 {
		log.Print("No missing CIDRs to add")
		return
	}
	log.Printf("Adding missing CIDRs %s to secgroup %s", CIDRs, secGroupID)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	for _, cidr := range CIDRs {
		svc := ec2.New(sess)
		_, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: &secGroupID,
			IpPermissions: []*ec2.IpPermission{
				(&ec2.IpPermission{}).SetIpProtocol("tcp").
					SetFromPort(80).
					SetToPort(443).
					SetIpRanges([]*ec2.IpRange{
						{CidrIp: aws.String(cidr)},
					}),
			},
		})
		exitIfError(fmt.Sprintf("ERROR adding %s", cidr), err)
	}

}

func deleteObsoleteCIDRs(secGroupID string, CIDRs []string) {
	if len(CIDRs) == 0 {
		log.Print("No obsolete CIDRs to delete")
		return
	}
	log.Printf("Deleting obsolete CIDRs %s from secgroup %s", CIDRs, secGroupID)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	for _, cidr := range CIDRs {
		svc := ec2.New(sess)
		_, err := svc.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: &secGroupID,
			IpPermissions: []*ec2.IpPermission{
				(&ec2.IpPermission{}).SetIpProtocol("tcp").
					SetFromPort(80).
					SetToPort(443).
					SetIpRanges([]*ec2.IpRange{
						{CidrIp: aws.String(cidr)},
					}),
			},
		})
		exitIfError(fmt.Sprintf("ERROR removing %s", cidr), err)
	}
}
