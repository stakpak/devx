package main

import (
	"guku.io/devx"
	"guku.io/devx/transformers/compose"
)

devx.#Application & {
	components: {
		proxy: devx.#Service & {
			image: "nginx"
			ports: [
				{
					port:   8123
					target: 80
				},
			]
			env: APP_URL:
				"http://\(app.outputs.host):\(app.ports[0].port)"
		}
		app: devx.#Service & {
			image: "app:v1"
			ports: [
				{
					port: 8080
				},
			]
			env: {
				DB_URL: db.outputs.url
			}
		}
		db: devx.#PostgresDB & {
			version:    "12.1"
			persistent: true
		}
	}
}

environments: {
	dev: {
		Service:    compose.#ComposeService
		PostgresDB: compose.#ComposePostgresDB
	}
}
