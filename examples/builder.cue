package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
	"guku.io/devx/v1/transformers/terraform"
	"guku.io/devx/v1/transformers/argocd"
)

builders: v1.#StackBuilder & {
	dev: {
		flows: [
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
		flows: [
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
