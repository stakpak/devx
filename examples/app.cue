package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	"guku.io/devx/v1/transformers/compose"
	"guku.io/devx/v1/transformers/database"
)

stack: v1.#Stack & {
	components: {
		app: {
			v1.#Component
			traits.#Workload
			traits.#Exposable

			$metadata: labels: application: "backend"

			endpoints: default: port: "8080"
			containers: default: {
				image: "app:v1"
				env: {
					POSTGRES_HOST:     db.endpoints.default.host
					POSTGRES_PORT:     db.endpoints.default.port
					POSTGRES_DATABASE: db.database.database
				}
			}
		}
		db: {
			v1.#Component
			traits.#Database
			database: {
				type:       "postgres"
				version:    "12.1"
				persistent: true
			}
		}
	}
}

builders: v1.#StackBuilder & {
	dev: flows: [
		{pipeline: [database.#AddDatabaseWorkload]},
		{pipeline: [database.#AddDatabaseVolume]},
		{pipeline: [compose.#AddComposeVolume]},
		{pipeline: [compose.#AddComposeService]},
		{pipeline: [compose.#ComposeExposeService]},
		{
			match: labels: application: "backend"
			pipeline: [compose.#ForwardPort & {
				args: endpoints: default: port: "8080"
			}]
		},
	]
}
