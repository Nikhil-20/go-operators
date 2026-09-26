package main

import (
	"context"
	"fmt"
	"k8s-watchdog/config"
	"k8s-watchdog/pkg/kube"
	"k8s-watchdog/watcher"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	fmt.Println("K8s watch dog starting...")

	//signal.Notify()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config:%v", err)
	}

	fmt.Printf("watchign resources : %v in namespaces:%v\n", cfg.Watch.Resources, cfg.Watch.Namespaces)

	// temporary verification of the k8s connection
	clientset, err := kube.GetClientSet()

	if err != nil {
		log.Fatalf("failed to create kubernetesclient :%v", err)
	}

	// serverVersion, err := clientset.Discovery().ServerVersion()

	// if err != nil {
	// 	log.Fatalf("Failed to get Kubernetes server version via clientset: %v", err)
	// }
	// fmt.Printf("Successfully connected to Kubernetes API server version: %s\n", serverVersion.GitVersion)

	// fmt.Println("--- Connection Verification Complete ---\n")
	// // --- END TEMPORARY VERIFICATION CODE ---

	watcher.WatchPod(ctx, clientset, cfg.Watch.Namespaces)

	<-ctx.Done()
	fmt.Println("\n[!] Shutdown signal received. Cleaning up streams and exiting...")
}
