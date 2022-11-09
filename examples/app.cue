package main

import (
	"guku.io/devx"
	devxc "guku.io/devx/components"
)

devx.#Application & {
	components: {
		proxy: devxc.#Service & {
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
		app: devxc.#Service & {
			image: "app:v1"
			ports: [
				{
					port: 8080
				},
			]
			env: {
				PGDB_URL: db.url
				MYDB_URL: mydb.url
			}
			volumes: [
				{
					source: "bla"
					target: "/tmp/bla"
				},
			]
		}
		db: devxc.#PostgresDB & {
			version:    "12.1"
			persistent: true
		}
		mydb: devxc.#MysqlDB & {
			version:    "8"
			persistent: true
		}
	}
}
