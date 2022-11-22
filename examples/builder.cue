package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
	"guku.io/devx/v1/transformers/terraform"
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
}
