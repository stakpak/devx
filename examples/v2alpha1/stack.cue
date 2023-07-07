package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/traits"
	"stakpak.dev/devx/v2alpha1"
	"stakpak.dev/devx/v2alpha1/environments"
)

// stack builder
builders: v2alpha1.#Environments & {
	dev: environments.#Compose
}

// common
stack: v1.#Stack & {
	components: {
		db: {
			traits.#Database
			traits.#Secret
			database: {
				name:     "prod"
				version:  "13.7"
				username: "root"
				password: secrets.dbPassword
			}
			secrets: dbPassword: name: "prod-db-password"
		}
		kafka: {
			traits.#Kafka
			traits.#Secret
			kafka: name: "prod"
			secrets: kafkaCreds: name: "kafka-user-c"
		}
		app2: {
			traits.#Workload
			traits.#Exposable
			containers: default: {
				image: "app"
				resources: requests: {
					cpu:    "256m"
					memory: "512M"
				}
				env: {
					KAFKA_URLS:   kafka.kafka.bootstrapServers
					KAFKA_SECRET: kafka.secrets.kafkaCreds
					DB_HOST:      db.database.host
					DB_PORT:      "\(db.database.port)"
					DB_NAME:      db.database.database
					DB_USERNAME:  db.database.username
					DB_PASSWORD:  db.database.password
				}
			}
			endpoints: default: ports: [
				{
					port:   8005
					target: 8000
				},
				{
					port: 8001
				},
			]
		}
	}
}
