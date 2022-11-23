package compose

import (
	"list"
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
		command?: [...string]
		volumes?: [...string]
	}
}

// add a compose service
#AddComposeService: v1.#Transformer & {
	$metadata: transformer: "AddComposeService"

	args: {}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Workload
		...
	}
	output: {
		endpoints: default: host: "\(input.$metadata.id)"
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				image:       input.containers.default.image
				environment: input.containers.default.env
				depends_on:  context.dependencies
				command:     list.Concat([
						input.containers.default.command,
						input.containers.default.args,
				])
				volumes: [
					for m in input.containers.default.mounts {
						_mapping: [
								if m.volume.local != _|_ {"\(m.volume.local):\(m.path)"},
								if m.volume.persistent != _|_ {"\(m.volume.persistent):\(m.path)"},
						][0]
						_suffix: [
								if m.readOnly {":ro"},
								if !m.readOnly {""},
						][0]
						"\(_mapping)\(_suffix)"
					},
				]
			}
		}
	}
}

// add a compose service
#AddComposeVolume: v1.#Transformer & {
	$metadata: transformer: "AddComposeVolume"

	args: {}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Volume
		...
	}
	output: {
		$resources: compose: _#ComposeResource & {
			for k, v in input.volumes {
				// only persistent volumes supported in compose
				if v.persistent != _|_ {
					volumes: "\(v.persistent)": null
				}
			}
		}
	}
}

// expose a compose service ports
#ExposeComposeService: v1.#Transformer & {
	$metadata: transformer: "ExposeComposeService"

	args: {}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Exposable
		...
	}
	output: {
		endpoints: default: host: "\(input.$metadata.id)"
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				ports: [
					for p in input.endpoints.default.ports {
						"\(p.port):\(p.target)"
					},
				]
			}
		}
	}
}

// add a compose service for a postgres database
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
