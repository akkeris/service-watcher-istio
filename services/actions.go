package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	utils "service-watcher-istio/utils"

	corev1 "k8s.io/api/core/v1"
)

type Spacespec struct {
	Name     string `json:"name"}`
	Internal bool   `json:"internal"}`
}

type HeaderOperationsspec struct {
	Set    map[string]string `json:"set,omitempty"`
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

type Headersspec struct {
	Request  HeaderOperationsspec `json:"request,omitempty"`
	Response HeaderOperationsspec `json:"response,omitempty"`
}

type Virtualservice struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Gateways []string   `json:"gateways"`
		Hosts    []string   `json:"hosts"`
		HTTP     []HTTPSpec `json:"http"`
	} `json:"spec"`
}

type HTTPSpec struct {
	Route   []Routespec `json:"route"`
	Headers Headersspec `json:"headers,omitempty"`
}
type Routespec struct {
	Destination struct {
		Host string `json:"host"`
		Port struct {
			Number int32 `json:"number"`
		} `json:"port"`
	} `json:"destination"`
}

func InstallGatewayVirtualservice(obj interface{}) {

	servicename := obj.(*corev1.Service).ObjectMeta.Name
	namespace := obj.(*corev1.Service).ObjectMeta.Namespace
	vsnamespace := "sites-system"
	port := obj.(*corev1.Service).Spec.Ports[0].Port
	InstallVirtualService(servicename, namespace, port, vsnamespace)

}

func DeleteGatewayVirtualservice(obj interface{}) {

	servicename := obj.(*corev1.Service).ObjectMeta.Name
	namespace := obj.(*corev1.Service).ObjectMeta.Namespace
	vsnamespace := "sites-system"
	DeleteVirtualservice(servicename, namespace, vsnamespace)

}

func InstallVirtualService(servicename string, namespace string, port int32, vsnamespace string) {

	appname := servicename + "-" + namespace
	if namespace == "default" {
		appname = servicename
	}
	var url string
	var gateway string
	internal := isInternal(namespace)
	if internal {
		url = appname + "." + utils.InsideDomain
		gateway = "apps-private"
	}
	if !internal {
		url = appname + "." + utils.DefaultDomain
		gateway = "apps-public"
	}
	var v Virtualservice
	v.APIVersion = "networking.istio.io/v1alpha3"
	v.Kind = "VirtualService"
	v.Metadata.Name = servicename + "-" + namespace
	v.Metadata.Namespace = vsnamespace
	v.Spec.Gateways = append(v.Spec.Gateways, gateway)
	v.Spec.Hosts = append(v.Spec.Hosts, url)
	var r Routespec
	r.Destination.Host = servicename + "." + namespace + ".svc.cluster.local"
	r.Destination.Port.Number = port
	var routes []Routespec
	routes = append(routes, r)
	var h HTTPSpec
	h.Route = routes
	if h.Headers.Response.Set == nil {
		h.Headers.Response.Set = make(map[string]string)
	}
	h.Headers.Response.Set["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
	v.Spec.HTTP = append(v.Spec.HTTP, h)

	virtualservice, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("VIRTUALSERVICE: " + string(virtualservice))

	req, err := http.NewRequest("POST", utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+vsnamespace+"/virtualservices", bytes.NewBuffer(virtualservice))
	if err != nil {
		fmt.Println("Error creating request")
		fmt.Println(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+utils.Kubetoken)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, doerr := client.Do(req)
	fmt.Printf("%+v\n", resp)
	if doerr != nil {
		fmt.Println("Do error")
		fmt.Println(doerr)
	}
	defer resp.Body.Close()
	fmt.Println("install virtual service response: " + resp.Status)
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		fmt.Println("Error installing virtual service")
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Unable to read error body")
		} else {
			fmt.Println(string(bodyBytes))
		}
	}
}

func DeleteVirtualservice(servicename string, namespace string, vsnamespace string) {
	req, err := http.NewRequest("DELETE", utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+vsnamespace+"/virtualservices/"+servicename+"-"+namespace, nil)
	if err != nil {
		fmt.Println("Error creating request")
		fmt.Println(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+utils.Kubetoken)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, doerr := client.Do(req)
	fmt.Printf("%+v\n", resp)
	if doerr != nil {
		fmt.Println("Do error")
		fmt.Println(doerr)
	}
	defer resp.Body.Close()
	fmt.Println("delete virtual service response: " + resp.Status)
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		fmt.Println("Error deleting virtual service")
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Unable to read error body")
		} else {
			fmt.Println(string(bodyBytes))
		}
	}
}

func isInternal(space string) bool {

	req, err := http.NewRequest("GET", utils.Regionapilocation+"/v1/space/"+space, nil)
	req.SetBasicAuth(utils.Regionapiusername, utils.Regionapipassword)
	if err != nil {
		fmt.Println(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	bb, err := ioutil.ReadAll(resp.Body)
	var spaceobject Spacespec
	uerr := json.Unmarshal(bb, &spaceobject)
	if uerr != nil {
		fmt.Println(uerr)
	}
	fmt.Printf("ISINTERNAL: %v\n", spaceobject.Internal)
	return spaceobject.Internal
}
