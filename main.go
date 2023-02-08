package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/krateoplatformops/patch-provider/internal/functions"
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

	restConfig, err := RESTConfigFromBytes(yml, "") //o.kubeconfigContext)
	if err != nil {
		panic(err)
	}

	cl, err := client.New(restConfig, client.Options{})
	if err != nil {
		panic(err)
	}

	//t := template.Must(template.New("test").
	//	Funcs(functions.Register(cl)).
	//	Parse(`{{ secret "kube-system" "bootstrap-token-abcdef" "data.expiration" | b64dec }}`))

	//t := template.Must(template.New("test").
	//	Funcs(functions.Register(cl)).
	//	Parse(`{{ api "v1" "ConfigMap" "" "foo" "metadata.labels" }}`))

	t := template.Must(template.New("test").
		Funcs(functions.Register(cl)).
		Parse(`{{ cm "" "foo" "metadata.labels" }}`))

	err = t.Execute(os.Stdout, nil)
	if err != nil {
		fmt.Println(err)
	}

	/*
		x := &unstructured.Unstructured{}
		x.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})

		err = cl.Get(context.TODO(), types.NamespacedName{
			Name:      "foo",
			Namespace: "default",
		}, x)
		//cm := corev1.ConfigMap{}
		//err = cl.Get(context.TODO(), types.NamespacedName{
		//	Name:      "foo",
		//	Namespace: "default",
		//}, &cm)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("%+v\n", u)

		//in, err := fieldpath.Pave(u.Object).GetValue("metadata.labels.hello")
		//if err != nil {
		//	panic(err)
		//}
		//fmt.Println("==>", in)

		y := x.DeepCopy()
		err = patching.Apply(v1alpha1.Patch{
			Spec: v1alpha1.PatchSpec{
				From: helpers.StringPtr("metadata.labels.hello"),
				To:   helpers.StringPtr("metadata.labels.patched-by"),
			},
		}, x, y)
		if err != nil {
			panic(err)
		}

		fmt.Println(cmp.Diff(x, y))
		/*
				cm.Labels["hello"] = "krateo"
				cm.Labels["patched-by"] = "pinco-pallo"

			applicator := resource.NewAPIPatchingApplicator(cl)
			if err := applicator.Apply(context.TODO(), y); err != nil {
				panic(err)
			}

			//gv, err := schema.ParseGroupVersion("krateo.io/v1alpha1")
			//if err != nil {
			//	panic(err)
			//}
	*/
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
