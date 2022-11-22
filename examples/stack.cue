package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		somechart: {
			v1.#Component
			traits.#Helm
			chart:     "hello-kubernetes-chart"
			url:       "https://somechart.github.io/my-charts/"
			version:   "0.1.0"
			namespace: "tata"
		}
		app: {
			v1.#Component
			traits.#Workload
			traits.#Exposable
			$metadata: labels: app: "app1"
			containers: default: {
				image: "app:v1"
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
			endpoints: default: {
				ports: [
					{
						port: 8080
					},
				]
			}
		}
		db: {
			v1.#Component
			traits.#Postgres
			version:    "12.1"
			persistent: true
		}
	}
}
