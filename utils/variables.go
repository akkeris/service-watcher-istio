
package utils

import (
	"os"
	"strings"
	"fmt"
	"k8s.io/client-go/rest"
)

var Kubetoken string
var DefaultDomain string
var InsideDomain string
var Kubernetesapiurl string
var Blacklist map[string]bool
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
