package main

import (
	"encoding/json"
	"fmt"
	"github.com/stackimpact/stackimpact-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	k8sconfig "service-watcher-istio/k8sconfig"
	services "service-watcher-istio/services"
	utils "service-watcher-istio/utils"
	"time"
)

func main() {
	utils.SetSecrets()
	if os.Getenv("PROFILE") == "true" {
		fmt.Println("Starting profiler...")
		_ = stackimpact.Start(stackimpact.Options{
			AgentKey:       "e4795df33a4002d0ddd7952dd8a8138c8312a449",
			AppName:        "Service Watcher",
			AppEnvironment: os.Getenv("CLUSTER"),
		})

	}

	k8sconfig.CreateConfig()
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	utils.Client = clientset.Core().RESTClient()
	listWatch := cache.NewListWatchFromClient(
		utils.Client, "services", "",
		fields.Everything())

	listWatch.ListFunc = func(options api.ListOptions) (runtime.Object, error) {
		return utils.Client.Get().Namespace("none").Resource("services").Do().Get()
	}
	listWatch.WatchFunc = func(options api.ListOptions) (watch.Interface, error) {
		return clientset.Core().Services(api.NamespaceAll).Watch(v1.ListOptions{})
	}

	_, controller := cache.NewInformer(
		listWatch, &corev1.Service{},
		time.Second*0, cache.ResourceEventHandlerFuncs{
			AddFunc:    printEventAdd,
			DeleteFunc: printEventDelete,
		},
	)
	fmt.Println("Watching for changes in Services....")
	controller.Run(wait.NeverStop)
}

func printEventAdd(obj interface{}) {
	_, isService := obj.(*corev1.Service)
	if isService {

		created := obj.(*corev1.Service).ObjectMeta.CreationTimestamp.Unix()
		now := v1.Now().Unix()

		diff := now - created
		if diff < 300 && !Blacklisted(obj.(*corev1.Service).ObjectMeta.Namespace) {
			fmt.Println("ADD")
			var err error

			_, err = json.Marshal(obj)
			if err != nil {
				fmt.Println(err)
				return
			}
			services.InstallGatewayVirtualservice(obj)
		}
	}
}

func Blacklisted(namespace string) bool {

	return utils.Blacklist[namespace]

}

func printEventDelete(obj interface{}) {
	fmt.Println("DELETE")
	_, isService := obj.(*corev1.Service)
	if isService {
		if !Blacklisted(obj.(*corev1.Service).ObjectMeta.Namespace) {
			var err error
			_, err = json.Marshal(obj)
			if err != nil {
				fmt.Println(err)
				return
			}
                        services.DeleteGatewayVirtualservice(obj)
		}
	}

}
