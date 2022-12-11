package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/transformers/compose"
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
