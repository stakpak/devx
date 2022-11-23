package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#ExposeComposeService & {},
				]
			},
			v1.#Flow & {
				pipeline: [
					compose.#AddComposeVolume & {},
				]
			},
		]
	}
}
