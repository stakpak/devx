package github

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#PipelineResource: {
	_#GitHubCISpec
	$metadata: labels: {
		driver: "github"
		type:   ""
	}
}

#AddCIPipeline: v1.#Transformer & {
	v1.#Component
	traits.#Workflow
	$metadata: _
	plan:      _#GitHubCISpec

	$resources: "\($metadata.id)-github": _#PipelineResource & {
		// for some reason the CUE resolver vails if we don't add $metadata again here
		$metadata: labels: {
			driver: "github"
			type:   ""
		}
		plan
	}
}
