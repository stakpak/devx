package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/transformers/argocd"
	terraform "stakpak.dev/devx/v1/transformers/terraform/helm"
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
