
package utils

import (
	"os"
	"strings"
	"fmt"
	"k8s.io/client-go/rest"
        vault "github.com/akkeris/vault-client"
)

var Kubetoken string
var DefaultDomain string
var InsideDomain string
var Kubernetesapiurl string
var Blacklist map[string]bool
var Client rest.Interface

func SetSecrets() {

	Kubetoken = vault.GetField(os.Getenv("KUBERNETES_TOKEN_VAULT_PATH"), "token")
	DefaultDomain = os.Getenv("DEFAULT_DOMAIN")
	InsideDomain = os.Getenv("INSIDE_DOMAIN")
	Kubernetesapiurl = os.Getenv("KUBERNETES_API_SERVER")
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
