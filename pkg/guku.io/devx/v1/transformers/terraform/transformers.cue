package terraform

import (
	"encoding/yaml"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#TerraformResource: {
	$metadata: labels: {
		driver: "terraform"
		type:   string
	}
}

// add a helm release
#AddHelmRelease: v1.#Transformer & {
	v1.#Component
	traits.#Helm
	$dependencies: [...string]
	$metadata: _
	namespace: string
	url:       _
	chart:     _
	version:   _
	values:    _
	$resources: terraform: {
		_#TerraformResource
		"$metadata": labels: type: "helm_release"
		resource: helm_release: "\($metadata.id)": {
			name:             $metadata.id
			"namespace":      namespace
			repository:       url
			"chart":          chart
			"version":        version
			create_namespace: true
			"values": [
				yaml.Marshal(values),
			]
		}
	}
}
