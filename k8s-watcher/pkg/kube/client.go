package kube

import (
	"fmt"
	"log"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetClientSet() (*kubernetes.Clientset, error) {

	conf, err := rest.InClusterConfig()

	if err == nil {
		log.Println("working with incluster kubernetes config")
		return kubernetes.NewForConfig(conf)
	}

	log.Println("using out of cluster config with kube/config for local testing")
	var kubeconfigpath string

	if home := homedir.HomeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	} else {
		return nil, fmt.Errorf("Error is getting the kube config path")
	}

	conf, err = clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconf from %s:%w", kubeconfigpath, err)
	}

	return kubernetes.NewForConfig(conf)
}
