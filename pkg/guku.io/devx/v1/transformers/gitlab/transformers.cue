package gitlab

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#PipelineResource: {
	#GitlabCISpec
	$metadata: labels: driver: "gitlab"
}

#AddCIPipeline: v1.#Transformer & {
	v1.#Component
	traits.#Workflow
	$metadata: _
	plan:      #GitlabCISpec

	$resources: "\($metadata.id)-gitlab": _#PipelineResource & {
		plan
	}
}
