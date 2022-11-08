package main

import (
	"guku.io/devx"
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
				"http://\(app.host):\(app.ports[0].port)"
		}
		app: devx.#Service & {
			image: "app:v1"
			ports: [
				{
					port: 8080
				},
			]
			env: {
				DB_URL: db.url
			}
		}
		db: devx.#PostgresDB & {
			version:    "12.1"
			persistent: true
		}
	}
}
