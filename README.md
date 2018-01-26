# cdn-securitygroup-sync

Note: This is a fork from https://github.com/schnoddelbotz/cdn-securitygroup-sync I just made some
adjustments after AWS now supports go lambdas in a native way. All credits to schnoddelbotz and his work.

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
  -edgegrid-access-token string
    	Akamai edgegrid access token
  -edgegrid-client-secret string
    	Akamai edgegrid client secret
  -edgegrid-client-token string
    	Akamai edgegrid client token
  -edgegrid-host string
    	Akamai host
  -list-ss-ids
    	List Akamai siteshield IDs and quit
  -sgid string
    	AWS security group ID
  -ssid int
    	Akamai siteshield ID
  -version
    	Print version and quit
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

The lambda approach assumes that you pass in all runtime configuration in either plaintext or
encrypt it via KMS. If parameters are passed in as plaintext use something like this: (example from
a terraform configuration):

```
  environment {
    variables = {
      AWS_SECGROUP_ID = "${var.security_group_id}"
      AKAMAI_SSID = "${var.akamai_ssid}"
      AKAMAI_EDGEGRID_HOST          = ""
      AKAMAI_EDGEGRID_CLIENT_TOKEN  = ""
      AKAMAI_EDGEGRID_CLIENT_SECRET = ""
      AKAMAI_EDGEGRID_ACCESS_TOKEN  = ""
      CSS_ARGS = "-add-missing,-delete-obsolete"
    }
  }
```

If KMS is used to encrypt credentials, please use the following config block:

```
  environment {
    variables = {
      AWS_SECGROUP_ID = "${var.security_group_id}"
      KMS_AKAMAI_SSID = "${var.akamai_ssid}"
      KMS_AKAMAI_EDGEGRID_HOST          = ""
      KMS_AKAMAI_EDGEGRID_CLIENT_TOKEN  = ""
      KMS_AKAMAI_EDGEGRID_CLIENT_SECRET = ""
      KMS_AKAMAI_EDGEGRID_ACCESS_TOKEN  = ""
      CSS_ARGS = "-add-missing,-delete-obsolete"
    }
  }
```


## deploy the lambda handler

We are using terraform for our deployment. Therefore we have a main.tf file included in this repository. Feel free
to copy and change this terraform file and adjust it for your needs.


# license

MIT.

Use cdn-securitygroup-sync at your own risk!

This project includes these 3rd party libraries to do its job:

- [AkamaiOPEN-edgegrid-golang](https://github.com/akamai/AkamaiOPEN-edgegrid-golang)
- [AWS golang SDK](github.com/aws/aws-sdk-go/aws)
- [aws-lambda-go-shim](https://github.com/eawsy/aws-lambda-go-shim)
