package main

import (
	"context"
	"fmt"
	"os"

	"github.com/krateoplatformops/patch-provider/apis"
	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/patching"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	defaultKubeconfig := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(defaultKubeconfig) == 0 {
		defaultKubeconfig = clientcmd.RecommendedHomeFile
	}

	yml, err := os.ReadFile(defaultKubeconfig)
	if err != nil {
		panic(err)
	}

	apis.AddToScheme(scheme.Scheme)

	restConfig, err := RESTConfigFromBytes(yml, "") //o.kubeconfigContext)
	if err != nil {
		panic(err)
	}

	cl, err := client.New(restConfig, client.Options{})
	if err != nil {
		panic(err)
	}

	cr := v1alpha1.Patch{}
	err = cl.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "sample-1"}, &cr)
	if err != nil {
		panic(err)
	}

	from, err := patching.FromObjectReference(context.TODO(), cl, &cr)
	if err != nil {
		panic(err)
	}

	to, err := patching.ToObjectReference(context.TODO(), cl, &cr)
	if err != nil {
		panic(err)
	}

	diff, err := patching.Diff(&cr, from, to)
	if err != nil {
		panic(err)
	}

	if len(diff) == 0 {
		fmt.Println("Patch not needed.")
		return
	}

	fmt.Println(diff)

	if err := patching.Patch(&cr, from, to); err != nil {
		panic(err)
	}

	if err := patching.Apply(context.TODO(), cl, to); err != nil {
		panic(err)
	}

}

func RESTConfigFromBytes(data []byte, withContext string) (*rest.Config, error) {
	config, err := clientcmd.Load(data)
	if err != nil {
		return nil, err
	}

	currentContext := config.CurrentContext
	if len(withContext) > 0 {
		currentContext = withContext
	}

	restConfig, err := clientcmd.NewNonInteractiveClientConfig(*config,
		currentContext, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	// Set QPS and Burst to a threshold that ensures the client doesn't generate throttling log messages
	restConfig.QPS = 20
	restConfig.Burst = 100

	return restConfig, nil
}
