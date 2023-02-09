package functions

import (
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Register2(kc client.Client) template.FuncMap {
	return template.FuncMap(map[string]any{
		// Kubernetes Resources
		"api":    fetchUnknown(kc),
		"cm":     fetchKnown(kc, "v1", "ConfigMap"),
		"secret": fetchKnown(kc, "v1", "Secret"),

		// Encoding:
		"b64enc": base64encode,
		"b64dec": base64decode,
	})
}

func Map() map[string]any {
	return map[string]any{
		// Encoding:
		"b64enc": base64encode,
		"b64dec": base64decode,
		"toJson": toJson,
	}
}
