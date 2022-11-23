package terraform

import (
	"encoding/yaml"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#TerraformResource: {
	$metadata: labels: driver: "terraform"
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
		$resources: terraform: {
			_#TerraformResource
			resource: helm_release: "\(input.$metadata.id)": {
				name:             input.$metadata.id
				namespace:        _namespace
				repository:       input.url
				chart:            input.chart
				version:          input.version
				create_namespace: true
				values: [
					yaml.Marshal(input.values),
				]
			}
		}
	}
}
