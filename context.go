package core

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// WebhookGlobalContext global context of webhooks
// this context will be used to share global configuration and objects between webhooks and server
type WebhookGlobalContext struct {
	config    *restclient.Config
	clientset *kubernetes.Clientset
}

var kubeconfig string
var masterURL string
var context *WebhookGlobalContext

func init() {
	home := os.Getenv("HOME")
	defaultPath := filepath.Join(home, ".kube", "config")
	flag.StringVar(&kubeconfig, "kubeconfig", defaultPath, "path to kubeconfig file")
	flag.StringVar(&masterURL, "masterURL", "", "master URL for kubernetes")
}

// InitializeGlobalContext initialize global context.
// DO NOT call this function, it will be called by server after `flag.Parse()`
func InitializeGlobalContext() error {
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	context = &WebhookGlobalContext{
		config:    config,
		clientset: clientset,
	}
	return nil
}

// GlobalContext return webhook global context
func GlobalContext() *WebhookGlobalContext { return context }

// GetRESTConfig get REST configuration that created from loading kubeconfig
func (context *WebhookGlobalContext) GetRESTConfig() *restclient.Config { return context.config }

// GetClientset get Clientset object from global context
func (context *WebhookGlobalContext) GetClientset() *kubernetes.Clientset { return context.clientset }
