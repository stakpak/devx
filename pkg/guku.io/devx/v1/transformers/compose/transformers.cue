package compose

import (
	"strings"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#ComposeResource: {
	$metadata: labels: driver: "compose"

	version: string | *"3"
	volumes: [string]: null
	services: [string]: {
		image: string
		depends_on?: [...string]
		ports?: [...string]
		environment?: [string]: string
		command?: string
		volumes?: [...string]
	}
}

#AddComposeService: v1.#Transformer & {
	$metadata: transformer: "AddComposeService"

	args: {}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Workload
		traits.#Exposable
		...
	}
	output: {
		endpoints: default: host: "\(input.$metadata.id)"
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				image: input.containers.default.image
				ports: [
					for p in input.endpoints.default.ports {
						"\(p.port):\(p.target)"
					},
				]
				environment: input.containers.default.env
				depends_on:  context.dependencies
				volumes: [
					for v in input.containers.default.volumes {
						if v.readOnly {
							"\(v.source):\(v.target):ro"
						}
						if !v.readOnly {
							"\(v.source):\(v.target)"
						}
					},
				]
			}
			for v in input.containers.default.volumes {
				if !strings.HasPrefix(v.source, ".") && !strings.HasPrefix(v.source, "/") {
					volumes: "\(v.source)": null
				}
			}
		}
	}
}

#AddComposePostgres: v1.#Transformer & {
	$metadata: transformer: "AddComposePostgres"

	args: {}
	context: {
		dependencies: [...string]
		_username: string @guku(generate)
		_password: string @guku(generate,secret)
	}
	input: {
		v1.#Component
		traits.#Postgres
		...
	}
	output: {
		host:     "\(input.$metadata.id)"
		username: context._username
		password: context._password
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				image: "postgres:\(input.version)-alpine"
				ports: [
					"\(input.port)",
				]
				if input.persistent {
					volumes: [
						"pg-data:/var/lib/postgresql/data",
					]
				}
				environment: {
					POSTGRES_USER:     context._username
					POSTGRES_PASSWORD: context._password
					POSTGRES_DB:       input.database
				}
				depends_on: context.dependencies
			}
			if input.persistent {
				volumes: "pg-data": null
			}
		}
	}
}
