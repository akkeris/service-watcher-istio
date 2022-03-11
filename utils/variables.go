package utils

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/rest"
)

var Kubetoken string
var DefaultDomain string
var InsideDomain string
var Kubernetesapiurl string
var Blacklist map[string]bool
var IgnoreList map[string]bool
var Client rest.Interface
var Regionapilocation string
var Regionapiusername string
var Regionapipassword string

func SetSecrets() {

	Kubetoken = os.Getenv("KUBERNETES_TOKEN")
	DefaultDomain = os.Getenv("DEFAULT_DOMAIN")
	InsideDomain = os.Getenv("INSIDE_DOMAIN")
	Kubernetesapiurl = os.Getenv("KUBERNETES_API_SERVER")
	Regionapiusername = os.Getenv("REGIONAPI_USERNAME")
	Regionapipassword = os.Getenv("REGIONAPI_PASSWORD")
	Regionapilocation = os.Getenv("REGIONAPI_URL")

	initBlacklist()
	initIgnoreList()
}

func initBlacklist() {
	Blacklist = make(map[string]bool)
	blackliststring := os.Getenv("NAMESPACE_BLACKLIST")
	blacklistslice := strings.Split(blackliststring, ",")
	for _, element := range blacklistslice {
		Blacklist[element] = true
	}
	keys := make([]string, 0, len(Blacklist))
	for k := range Blacklist {
		keys = append(keys, k)
	}

	fmt.Printf("Setting blacklist to %v\n", strings.Join(keys, ","))

}

func initIgnoreList() {
	IgnoreList = make(map[string]bool)
	ignoreliststring := os.Getenv("IGNORE_LABELS")
	ignorelistslice := strings.Split(ignoreliststring, ",")
	for _, element := range ignorelistslice {
		IgnoreList[element] = true
	}
	keys := make([]string, 0, len(IgnoreList))
	for k := range IgnoreList {
		keys = append(keys, k)
	}
	fmt.Printf("Setting ignoreList to %v\n", strings.Join(keys, ","))
}
