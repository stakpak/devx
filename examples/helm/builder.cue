package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/argocd"
	terraform "guku.io/devx/v1/transformers/terraform/helm"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					argocd.#AddHelmRelease & {helm: namespace: string | *"default"},
				]
			},
		]
	}
	prod: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					terraform.#AddHelmRelease & {helm: namespace: "somethingelse"},
				]
			},
		]
	}
}
