package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
        corev1 "k8s.io/api/core/v1"
        utils "service-watcher-istio/utils"
)

type Spacespec struct {
        Name     string `json:"name"}`
        Internal bool   `json:"internal"}`
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
	Route []Routespec `json:"route"`
}
type Routespec struct {
	Destination struct {
		Host string `json:"host"`
		Port struct {
			Number int `json:"number"`
		} `json:"port"`
	} `json:"destination"`
}

type Gateway struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Selector struct {
			Istio string `json:"istio"`
		} `json:"selector"`
		Servers []Server `json:"servers"`
	} `json:"spec"`
}

type Server struct {
	Hosts []string `json:"hosts"`
	Port  struct {
		Name     string `json:"name"`
		Number   int    `json:"number"`
		Protocol string `json:"protocol"`
	} `json:"port"`
	TLS struct {
		CredentialName    string `json:"credentialName,omitempty"`
		Mode              string `json:"mode,omitempty"`
		PrivateKey        string `json:"privateKey,omitempty"`
		ServerCertificate string `json:"serverCertificate,omitempty"`
		HttpsRedirect     bool   `json:"httpsRedirect,omitempty"`
	} `json:"tls,omitempty"`
}


func InstallGatewayVirtualservice(obj interface{}) {


        servicename := obj.(*corev1.Service).ObjectMeta.Name
        namespace := obj.(*corev1.Service).ObjectMeta.Namespace
        port :=   80
	InstallGateway(servicename, namespace)
	InstallVirtualService(servicename, namespace, port)

}

func DeleteGatewayVirtualservice(obj interface{}) {


        servicename := obj.(*corev1.Service).ObjectMeta.Name
        namespace := obj.(*corev1.Service).ObjectMeta.Namespace
        DeleteGateway(servicename, namespace)
        DeleteVirtualservice(servicename, namespace)

}
func InstallGateway(servicename string, namespace string) {
	appname := servicename + "-" + namespace
	if namespace == "default" {
		appname = servicename
	}
	var url string
	var pp string
	internal := isInternal(namespace)
	if internal {
		url = appname + "." +utils.InsideDomain
		pp = "private"
	}
	if !internal {
		url = appname + "." + utils.DefaultDomain
		pp = "public"
	}

	var g Gateway
	g.APIVersion = "networking.istio.io/v1alpha3"
	g.Kind = "Gateway"
	g.Metadata.Name = servicename + "-gateway"
	g.Metadata.Namespace = namespace
	g.Spec.Selector.Istio = "apps-" + pp + "-ingressgateway"
	var s1 Server
	var s2 Server
	s1.Hosts = append(s1.Hosts, url)
	s1.Port.Name = "https-" + appname
	s1.Port.Number = 443
	s1.Port.Protocol = "HTTPS"
	s1.TLS.CredentialName = "apps-" + pp + "-certificate"
	s1.TLS.Mode = "SIMPLE"
	s1.TLS.PrivateKey = "/etc/istio/apps-public-certificate/tls.key"
	s1.TLS.ServerCertificate = "/etc/istio/apps-public-certificate/tls.crt"
	g.Spec.Servers = append(g.Spec.Servers, s1)

	s2.Hosts = append(s2.Hosts, url)
	s2.Port.Name = "http"
	s2.Port.Number = 80
	s2.Port.Protocol = "HTTP"
	s2.TLS.HttpsRedirect = true
	g.Spec.Servers = append(g.Spec.Servers, s2)

	gateway, err := json.Marshal(g)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("GATEWAY: " + string(gateway))

	req, err := http.NewRequest("POST", "https://"+utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+namespace+"/gateways", bytes.NewBuffer(gateway))
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
	fmt.Println("install gateway response: " + resp.Status)
}

func DeleteGateway(servicename string, namespace string){
        req, err := http.NewRequest("DELETE", "https://"+utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+namespace+"/gateways/"+servicename+"-gateway", nil)
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
        fmt.Println("delete gateway response: " + resp.Status)
}
func InstallVirtualService(servicename string, namespace string, port int) {

	appname := servicename + "-" + namespace
	if namespace == "default" {
		appname = servicename
	}
	var url string
	internal := isInternal(namespace)
	if internal {
		url = appname + "." + utils.InsideDomain
	}
	if !internal {
		url = appname + "." + utils.DefaultDomain
	}
	var v Virtualservice
	v.APIVersion = "networking.istio.io/v1alpha3"
	v.Kind = "VirtualService"
	v.Metadata.Name = servicename
	v.Metadata.Namespace = namespace
	v.Spec.Gateways = append(v.Spec.Gateways, servicename+"-gateway")
	v.Spec.Hosts = append(v.Spec.Hosts, url)
	var r Routespec
	r.Destination.Host = servicename + "." + namespace + ".svc.cluster.local"
	r.Destination.Port.Number = port
	var routes []Routespec
	routes = append(routes, r)
	var h HTTPSpec
	h.Route = routes
	v.Spec.HTTP = append(v.Spec.HTTP, h)

	virtualservice, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("VIRTUALSERVICE: " + string(virtualservice))

	req, err := http.NewRequest("POST", "https://"+utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+namespace+"/virtualservices", bytes.NewBuffer(virtualservice))
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
}

func DeleteVirtualservice(servicename string, namespace string){
        req, err := http.NewRequest("DELETE", "https://"+utils.Kubernetesapiurl+"/apis/networking.istio.io/v1alpha3/namespaces/"+namespace+"/virtualservices/"+servicename, nil)
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
  }
func AddLabel(namespace string) {


	type SpacePatch struct {
		Metadata struct {
			Labels struct {
				IstioInjection string `json:"istio-injection"`
			} `json:"labels"`
		} `json:"metadata"`
	}

	var sp SpacePatch
	sp.Metadata.Labels.IstioInjection = "enabled"

	spacepatch, err := json.Marshal(sp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("SPACEPATCH: " + string(spacepatch))

	req, err := http.NewRequest("PATCH", "https://"+utils.Kubernetesapiurl+"/api/v1/namespaces/"+namespace, bytes.NewBuffer(spacepatch))
	if err != nil {
		fmt.Println("Error creating request")
		fmt.Println(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+utils.Kubetoken)
	req.Header.Add("Content-Type", "application/merge-patch+json")
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
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(bodybytes))
	fmt.Println("install label response: " + resp.Status)
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
        fmt.Printf("ISINTERNAL: %v\n",spaceobject.Internal)
        return spaceobject.Internal
}

