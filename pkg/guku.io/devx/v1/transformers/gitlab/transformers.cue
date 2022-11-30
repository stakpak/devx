package gitlab

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#PipelineResource: {
	_#GitlabCISpec
	$metadata: labels: {
		driver: "gitlab"
		type:   ""
	}
}

#AddCIPipeline: v1.#Transformer & {
	v1.#Component
	traits.#Workflow
	$metadata: _
	plan:      _#GitlabCISpec

	$resources: "\($metadata.id)-gitlab": _#PipelineResource & {
		plan
	}
}
