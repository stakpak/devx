package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/transformers/compose"
)

builders: v1.#StackBuilder & {
	dev: {
		mainflows: [
			{
				pipeline: [compose.#AddComposeService]
			},
			{
				pipeline: [compose.#ExposeComposeService]
			},
			{
				pipeline: [compose.#AddComposePostgres]
			},
		]
	}
}
