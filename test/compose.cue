package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
)

_exposable: v1.#TestCase & {
	$metadata: test: "exposable"

	transformer: compose.#ExposeComposeService
	input: {
		$metadata: id: "obi"
		endpoints: default: ports: [
			{
				port:   8080
				target: 80
			},
		]
	}
	output: _

	expect: {
		endpoints: default: host: "obi"
		$resources: compose: services: obi: ports: ["8080:80"]
	}
	assert: {
		"host is concrete": (output.endpoints.default.host & "123") == _|_
	}
}
