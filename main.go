package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	//"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	listSSIDs      bool
	addMissing     bool
	deleteObsolete bool
	acknowledge    bool
	useCloudflare  bool
	printVersion   bool
	mySSID         int
	mySecGroup     string

	edgegridHost   string
	edgegridClientToken string
	edgegridClientSecret string
	edgegridAccessToken string

	// AppVersion is set at compile time
	AppVersion = "0.0.0-dev"
)

type Request struct {
	ID        float64 `json:"id"`
	Value     string  `json:"value"`
}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func main() {
	if 1 < len(os.Args) {
		// command line execution: parse command line flags
		parseFlags()
		// run() will also be invoked by lambda handler
		run()
	} else {
		lambda.Start(Handler)
	}
}

func Handler(request Request) (Response, error) {

	//log.SetFlags(0)
	// get runtime configuration from parameter store
	parseFlags()
	//ssmGet(os.Getenv("SSM_SOURCE"))
	// run() will also be invoked by lambda handler
	run()

	return Response{
		Message: fmt.Sprintf("Processed request ID %f", request.ID),
		Ok:      true,
	}, nil
}

func run() {
	config := getAkamaiConfig(edgegridHost, edgegridClientToken, edgegridClientSecret, edgegridAccessToken)
	if printVersion {
		println(version())
		os.Exit(0)
	}
	if listSSIDs {
		printSSIDs(config)
	}
	if mySSID == 0 && !useCloudflare {
		exitErrorf("Required argument -ssid / environment variable AKAMAI_SSID missing")
	}
	if mySecGroup == "" {
		exitErrorf("Required argument -sgid / environment variable AWS_SECGROUP_ID missing")
	}

	log.Print("cdn-securitygroup-sync starting...")

	// get AWS security group CIDRs
	sgCidrs := getSecGroupCIDRs(mySecGroup)

	ssCidrs := make(map[string]struct{})
	var ssMap siteShieldMap
	if useCloudflare {
		for _, cidr := range getCloudflareCIDRs() {
			ssCidrs[cidr] = struct{}{}
		}
	} else {
		// get Akamai siteshield data
		ssMap = getSiteshieldMap(config, mySSID)
		for _, cidr := range ssMap.CurrentCidrs {
			ssCidrs[cidr] = struct{}{}
		}
		// add proposed (i.e. non-acknowledged) CIDRs
		for _, cidr := range ssMap.ProposedCidrs {
			if _, ok := ssCidrs[cidr]; !ok {
				ssCidrs[cidr] = struct{}{}
			}
		}
	}

	// compare current with desired state
	missing := findMissingCidrs(sgCidrs, ssCidrs)
	obsolete := findObsoleteCidrs(sgCidrs, ssCidrs)

	// apply changes if requested
	if addMissing {
		addMissingCIDRs(mySecGroup, missing)
	}
	if deleteObsolete {
		deleteObsoleteCIDRs(mySecGroup, obsolete)
	}
	if ssMap.Acknowledged == false && !useCloudflare {
		if acknowledge {
			acknowledgeCIDRs(config, mySSID)
		} else {
			log.Print("Current Akamai CIDRs NOT acknowledged -- use -acknowledge to do so!")
		}
	} else if !useCloudflare {
		log.Print("Current Akamai CIDRs are acknowledged -- all good!")
	}

	log.Print("cdn-securitygroup-sync completed.")
}

func findKmsArg(envName string) string {
	envValue := os.Getenv("KMS_" + envName)
	if envValue != "" {
		retval, err := kmsDecrypt(envValue)
		fmt.Println(err)
		return retval
	}
	return os.Getenv(envName)
}

func parseFlags() {
	flag.BoolVar(&printVersion, "version", false, "Print version and quit")
	flag.BoolVar(&listSSIDs, "list-ss-ids", false, "List Akamai siteshield IDs and quit")
	flag.BoolVar(&useCloudflare, "cloudflare", false, "Use Cloudflare instead of Akamai")
	flag.BoolVar(&addMissing, "add-missing", false, "Add missing CIDRs to AWS security group")
	flag.BoolVar(&deleteObsolete, "delete-obsolete", false, "Delete obsolete CIDRs from AWS security group")
	flag.BoolVar(&acknowledge, "acknowledge", false, "Acknowledge updated CIDRs on Akamai")

	envSSID, _ := strconv.Atoi(findKmsArg("AKAMAI_SSID"))
	flag.IntVar(&mySSID, "ssid", envSSID, "Akamai siteshield ID")
	flag.StringVar(&mySecGroup, "sgid", findKmsArg("AWS_SECGROUP_ID"), "AWS security group ID")

	flag.StringVar(&edgegridHost, "edgegrid-host", findKmsArg("AKAMAI_EDGEGRID_HOST"), "Akamai host")
	flag.StringVar(&edgegridClientToken, "edgegrid-client-token", findKmsArg("AKAMAI_EDGEGRID_CLIENT_TOKEN"), "Akamai edgegrid client token")
	flag.StringVar(&edgegridClientSecret, "edgegrid-client-secret", findKmsArg("AKAMAI_EDGEGRID_CLIENT_SECRET"), "Akamai edgegrid client secret")
	flag.StringVar(&edgegridAccessToken, "edgegrid-access-token", findKmsArg("AKAMAI_EDGEGRID_ACCESS_TOKEN"), "Akamai edgegrid access token")


	cssArgs := os.Getenv("CSS_ARGS")
	if cssArgs != "" {
		parseLambdaFlags(cssArgs)
	}

	flag.Parse()
}

func findMissingCidrs(sg map[string]struct{}, ss map[string]struct{}) []string {
	// walk over SS cidrs and find thoses missing in sec group
	var list []string
	for cidr := range ss {
		if _, ok := sg[cidr]; !ok {
			log.Printf("Missing CIDR - exists in SS but not in SG: %s", cidr)
			list = append(list, cidr)
		}
	}
	if len(list) == 0 {
		log.Print("No missing CIDRs found -- all good!")
	}
	return list
}

func findObsoleteCidrs(sg map[string]struct{}, ss map[string]struct{}) []string {
	// walk over SG and find cidrs that dont exist in SS
	var list []string
	for cidr := range sg {
		if _, ok := ss[cidr]; !ok {
			log.Printf("Obsolete CIDR - exists in SG but not in SS: %s", cidr)
			list = append(list, cidr)
		}
	}
	if len(list) == 0 {
		log.Print("No obsolete CIDRs found -- all good!")
	}
	return list
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func exitIfError(msg string, e error) {
	if e != nil {
		exitErrorf(msg, "ERROR: %s, %v\n", msg, e)
	}
}

func version() string {
	if len(AppVersion) == 0 {
		AppVersion = "0.0.0-dev"
	}
	return AppVersion
}
