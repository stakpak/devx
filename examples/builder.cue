package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
	"guku.io/devx/v1/transformers/terraform"
	"guku.io/devx/v1/transformers/argocd"
	"guku.io/devx/v1/transformers/generic"
)

builders: dev: preflows: [
	v1.#Flow & {
		match: labels: {
			app: "app1"
		}
		pipeline: [
			generic.#AddExtraEnv & {
				args: env: canary: "canary"
			},
		]
	},
]

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
					compose.#ExposeComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#AddComposePostgres & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					terraform.#AddHelmRelease & {},
				]
			},
		]
	}
	dev2: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
					compose.#ExposeComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#AddComposePostgres & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					argocd.#AddHelmRelease & {},
				]
			},
		]
	}
}
