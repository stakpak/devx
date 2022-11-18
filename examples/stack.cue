package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		app: {
			v1.#Component
			traits.#Workload
			traits.#Exposable
			image: "app:v1"
			ports: [
				{
					port: 8080
				},
			]
			env: {
				PGDB_URL: db.url
			}
			volumes: [
				{
					source: "bla"
					target: "/tmp/bla"
				},
			]
		}
		db: {
			v1.#Component
			traits.#Postgres
			version:    "12.1"
			persistent: true
		}
	}
}
	