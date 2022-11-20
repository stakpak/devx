package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	"guku.io/devx/v1/transformers/compose"
)

builders: v1.#StackBuilder & {
	dev: {
		additionalComponents: {
			observedb: {
				v1.#Component
				traits.#Postgres
				version:    "12.1"
				persistent: true
			}
		}
		flows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#AddComposePostgres & {},
				]
			},
		]
	}
}
