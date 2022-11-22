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
		defaultNamespace: string
	}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Helm
		...
	}
	output: {
		namespace: input.namespace
		$resources: terraform: {
			_#TerraformResource
			resource: helm_release: "\(input.$metadata.id)": {
				name:             input.$metadata.id
				namespace:        input.namespace
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
