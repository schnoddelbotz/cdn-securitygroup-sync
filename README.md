# cdn-securitygroup-sync

Automates sync of AWS security groups with your CDN provider's CIDRs - currently
[Akamai Siteshield](https://community.akamai.com/community/cloud-security/blog/2016/11/15/list-of-ipscidrs-and-ports-on-the-akamai-network-that-may-contact-customers-origin-when-siteshield-is-enabled) 
and [Cloudflare](https://www.cloudflare.com/ips/) are supported.
Does basically the same job as [SSSG-Ninja](https://github.com/jc1518/SSSG-Ninja)
(for Akamai) but...

- comes as a single, ready-to-use, stand-alone binary
- comes with a CloudFormation stack for simple deployment as a scheduled AWS Lambda function
- has no hard-coded configuration data (like [this](https://github.com/jc1518/SSSG-Ninja/issues/2)
  or [that](https://github.com/jc1518/SSSG-Ninja/blob/6ba368a618a3bc667c59f3356d38c71f6c93efc6/securitygroup/__init__.py#L13))

# build / install

`go get -v github.com/schnoddelbotz/cdn-securitygroup-sync` to build
or grab a binary from the [releases page](../../releases).

# CLI usage

```
Usage of cdn-securitygroup-sync:
  -acknowledge
      Acknowledge updated CIDRs on Akamai
  -add-missing
      Add missing CIDRs to AWS security group
  -cloudflare
      Use Cloudflare instead of Akamai
  -delete-obsolete
      Delete obsolete CIDRs from AWS security group
  -list-ss-ids
      List Akamai siteshield IDs and quit
  -sgid string
      AWS security group ID
  -ssid int
      Akamai siteshield ID
```

Security group (`-sgid`) can be specified via envrionment variable `AWS_SECGROUP_ID`, too.
SiteShield ID (`-ssid`) can be alternatively provided via `AKAMAI_SSID`. Additionally,
for Akamai, these specific API environment variables must be defined:

- `AKAMAI_EDGEGRID_HOST`
- `AKAMAI_EDGEGRID_CLIENT_TOKEN`
- `AKAMAI_EDGEGRID_CLIENT_SECRET`
- `AKAMAI_EDGEGRID_ACCESS_TOKEN`

By default, `cdn-securitygroup-sync` will only list missing and obsolete CIDRs.
Arguments `-add-missing`, `-delete-obsolete` or `-acknowledge` have to be given 
explicitly to enable corresponding actions.

cdn-securitygroup-sync will create inbound rules on the given security group,
with a port range of 80-443, originating from CDN CIDRs. Any rules not using
the port range will remain untouched. You may rely on this behaviour for new
ELB/security group deployments: Create them with an inbound rule of
0.0.0.0/32, port range 80-443; upon first cdn-securitygroup-sync invocation
that rule will be removed and replaced by correct CDN CIDRs.

# lambda deployment

The lambda approach assumes that you store runtime configuration and credentials in parameter
store. To do so, [create a KMS key](http://docs.aws.amazon.com/kms/latest/developerguide/create-keys.html)
and refer to that key during stack deployment, as outlined below. You will also
have to provide a S3 bucket to store lambda code.

## put required entries into EC2 parameter store

The stack will create an IAM role that is granted KMS key access. Parameter store
entries will use a prefix (which defaults to `css`), which is used to restrict
access to the entries and allows to deploy multiple, independent instances of the lambda.

Using AWS CLI or AWS console, put these "secure string" parameters into parameter store
(assuming default prefix `css` in this example):

- `css_AWS_SECGROUP_ID`: The AWS EC2 security group to keep in sync (`sg-....`)
- `css_CSS_ARGS`: A comma-separated list of arguments for cdn-securitygroup-sync.
  Those arguments equal the command-line version of cdn-securitygroup-sync,
  i.e. to fully automate sync for Akamai, use `-add-missing,-delete-obsolete,-acknowledge`.
  To sync with Cloudflare (which doesn't require acknowledgement), use
  `-add-missing,-delete-obsolete,-cloudflare`.

If using Akamai, you will have to provide corresponding API credentials:

- `css_AKAMAI_SSID`: The SiteShield ID; can be obtained by using `-list-ss-ids` argument
- `css_AKAMAI_EDGEGRID_HOST`: Something like `xxxxxxx.luna.akamaiapis.net`
- `css_AKAMAI_EDGEGRID_CLIENT_TOKEN`
- `css_AKAMAI_EDGEGRID_CLIENT_SECRET`
- `css_AKAMAI_EDGEGRID_ACCESS_TOKEN`

There's no need to store any AWS credentials: The stack will create a policy
that grants the lambda required permissions to update the security group.

## deploy the lambda handler

There are two options for lambda deployment: Grab a pre-built lambda handler .zip
from the [releases](../../releases) page and upload it to your S3 bucket OR
build cdn-securitygroup-sync from source.

### variant 1 - deploy a pre-built release

- download the latest cdn-securitygroup-sync-lambda-....zip from [releases](../../releases) page
- upload the .zip to a S3 bucket (do not unzip!)
- deploy the lambda function using [cloudFormation stack](lambda/cf-stack.yaml),
  either via AWS console or by cloning this repository and running make:

```bash
make deploy-prebuilt AWS_REGION=eu-west-1 AWS_ACCOUNT_ID=123456... SSM_KEY_ID=abc-def \
        S3_BUCKET=my-little-bucket S3_KEY=path/to/cdn-securitygroup-sync-lambda-....zip
```

### variant 2 - deploy from source

Build dependencies: AWS-CLI, Docker, Go 1.8+, Make.

```bash
make deploy-source AWS_REGION=eu-west-1 AWS_ACCOUNT_ID=123456... SSM_KEY_ID=abc-def \
        S3_BUCKET=my-little-bucket
```

To just build and upload the lambda .zip to your S3 bucket named `my-little-bucket` for later (variant 1) usage:

```bash
# S3 key / destination path defaults to 'code/cdn-securitygroup-sync-$(VERSION).zip'
make S3_BUCKET=my-little-bucket
```

# license

MIT.

Use cdn-securitygroup-sync at your own risk!

This project includes these 3rd party libraries to do its job:

- [AkamaiOPEN-edgegrid-golang](https://github.com/akamai/AkamaiOPEN-edgegrid-golang)
- [AWS golang SDK](github.com/aws/aws-sdk-go/aws)
- [aws-lambda-go-shim](https://github.com/eawsy/aws-lambda-go-shim)
