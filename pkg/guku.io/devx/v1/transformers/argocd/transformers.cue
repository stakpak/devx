package argocd

import (
	"encoding/yaml"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	argoapp "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

_#ArgoCDApplicationResource: {
	$metadata: labels: driver: "kubernetes"
	argoapp.#Application
	spec: project: string | *"default"
}

// add a helm release
#AddHelmRelease: v1.#Transformer & {
	$metadata: transformer: "AddHelmRelease"

	args: {
		defaultNamespace:  string
		overrideNamespace: string
	}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Helm
		...
	}
	_namespace: [
			if (args.overrideNamespace & "*#?$**") == _|_ {args.overrideNamespace},
			if (input.namespace & "*#?$**") == _|_ {input.namespace},
			if (args.defaultNamespace & "*#?$**") == _|_ {args.defaultNamespace},
	][0]
	output: {
		namespace: _namespace
		$resources: "\(input.$metadata.id)": {
			_#ArgoCDApplicationResource
			kind:       "Application"
			apiVersion: "argoproj.io/v1alpha1"
			metadata: {
				name:      input.$metadata.id
				namespace: _namespace
				finalizers: [
					"resources-finalizer.argocd.argoproj.io",
				]
			}
			spec: {
				source: {

					chart: input.chart

					repoURL:        input.url
					targetRevision: input.version

					helm: {
						releaseName: input.$metadata.id
						values:      yaml.Marshal(input.values)
					}
				}
				destination: {
					namespace: _namespace
				}

				syncPolicy: argoapp.#SyncPolicy & {
					automated: {
						prune:      bool | *true
						selfHeal:   bool | *true
						allowEmpty: bool | *false
					}
					syncOptions: [...string] | *[
							"CreateNamespace=true",
							"PrunePropagationPolicy=foreground",
							"PruneLast=true",
					]
					retry: limit: uint | *5
				}
			}
		}
	}
}
