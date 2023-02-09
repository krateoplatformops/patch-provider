package main

import (
	"context"
	"os"

	"github.com/krateoplatformops/patch-provider/apis"
	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	if cr.Spec.From == nil || cr.Spec.From.ObjectReference == nil {
		return
	}

	gvk, err := schema.ParseGroupVersion(helpers.StringOrDefault(cr.Spec.From.ObjectReference.ApiVersion, "v1"))
	if err != nil {
		panic(err)
	}

	from := &unstructured.Unstructured{}
	from.SetGroupVersionKind(gvk.WithKind(cr.Spec.From.ObjectReference.Kind))
	err = cl.Get(context.TODO(), types.NamespacedName{
		Name:      cr.Spec.From.ObjectReference.Name,
		Namespace: helpers.StringOrDefault(cr.Spec.From.ObjectReference.Namespace, "default"),
	}, from)
	if err != nil {
		panic(err)
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

/*
func executeTemplate(kc client.Client, gist string) error {
	buf := bytes.NewBufferString("")
	tpl := template.New(cr.GetName()).Funcs(functions.Register(cl))
	tpl, err = tpl.Parse(*cr.Spec.From)
	if err != nil {
		panic(err)
	}

	err = tpl.Execute(os.Stdout, nil)
	if err != nil {
		fmt.Println(err)
	}
}
*/
