package watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func WatchPod(ctx context.Context, clientset *kubernetes.Clientset, namespaces []string) {

	if clientset == nil {
		log.Fatal("kube clientset is null..")
	}

	log.Println("Starting the kube pod watch")

	for _, ns := range namespaces {

		go func(namespace string) {
			log.Printf("Setting up watch for namespace: %s\n", namespace)

			factory := informers.NewSharedInformerFactoryWithOptions(
				clientset,
				time.Minute,
				informers.WithNamespace(namespace),
				informers.WithTweakListOptions(
					func(opt *metav1.ListOptions) {
						opt.FieldSelector = fields.Everything().String()
					}),
			)

			informer := factory.Core().V1().Pods().Informer()

			informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					pod := obj.(*corev1.Pod)
					fmt.Printf("Pod added: %s/%s\n", pod.Namespace, pod.Name)
				},
				DeleteFunc: func(obj interface{}) {
					pod := obj.(*corev1.Pod)
					fmt.Printf("[-] Pod deleted from %s: %s\n", namespace, pod.GetName())
				},
			})

			//stopch := make(chan struct{})

			//defer close(stopch)

			go factory.Start(ctx.Done())

			//factory.WaitForCacheSync(stopch)
			if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
				log.Printf("Cache sync aborted for namespace: %s\n", namespace)
				return
			}

			log.Printf("Cache synced for namespace: %s. Ready to watch events.\n", namespace)

			<-ctx.Done()

			log.Printf("Stopped watching namespace: %s\n", namespace)

		}(ns)
	}

	//select {}
}
