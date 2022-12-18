package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/kubernetes"
)

_hpa: v1.#TestCase & {
	$metadata: test: "hpa"

	transformer: kubernetes.#AddHPA & {
		replicas: max: 10
		hpaMetrics: [
			{
				type: "Resource"
				resource: {
					name: "cpu"
					target: {
						type:               "Utilization"
						averageUtilization: 60
					}
				}
			},
			{
				type: "Resource"
				resource: {
					name: "memory"
					target: {
						type:               "Utilization"
						averageUtilization: 60
					}
				}
			},
		]
	}
	input: {
		$metadata: id: "obi"
		replicas: min: 2
		$resources: "obi-deployment": {
			kind:       "Deployment"
			apiVersion: "apps/v1"
			metadata: name: "obi"
		}
	}
	output: _

	expect: {
		$resources: "obi-hpa": {
			metadata: {
				name: "obi"
				labels: app: "abi"
			}
			spec: {
				scaleTargetRef: {
					name:       "obi"
					kind:       "Deployment"
					apiVersion: "apps/v1"
				}
				minReplicas: 2
				maxReplicas: 10
				metrics:     transformer.hpaMetrics
			}
		}
	}
	assert: {
		"target target is concrete": (output.$resources["obi-hpa"].spec.scaleTargetRef & {
			name:       "123"
			kind:       "123"
			apiVersion: "123"
		} ) == _|_
		"min is concrete": (output.$resources["obi-hpa"].spec.minReplicas & 1000) == _|_
		"max is concrete": (output.$resources["obi-hpa"].spec.maxReplicas & 1000) == _|_
	}
}
