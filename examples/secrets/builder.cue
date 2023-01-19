package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
	tfaws "guku.io/devx/v1/transformers/terraform/aws"
	k8s "guku.io/devx/v1/transformers/kubernetes"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [compose.#AddComposeService]
			},
			v1.#Flow & {
				pipeline: [compose.#ExposeComposeService]
			},
			v1.#Flow & {
				pipeline: [compose.#AddComposeVolume]
			},
			{
				// allow secrets to not be fulfilled in strict mode
				match: traits: Secret: null
			},
		]
	}
	prod: {
		mainflows: [
			v1.#Flow & {
				pipeline: [k8s.#AddDeployment]
			},
			v1.#Flow & {
				pipeline: [k8s.#AddService]
			},
			v1.#Flow & {
				pipeline: [k8s.#AddWorkloadVolumes]
			},
			v1.#Flow & {
				pipeline: [tfaws.#AddSSMSecretParameter]
			},
		]
	}
}
