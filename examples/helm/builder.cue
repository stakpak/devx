package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/argocd"
	"guku.io/devx/v1/transformers/terraform"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					argocd.#AddHelmRelease & {namespace: string | *"default"},
				]
			},
		]
	}
	prod: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					terraform.#AddHelmRelease & {namespace: "somethingelse"},
				]
			},
		]
	}
}
