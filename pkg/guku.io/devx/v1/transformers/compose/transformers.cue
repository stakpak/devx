package compose

import (
	"list"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#ComposeResource: {
	$metadata: labels: {
		driver: "compose"
		type:   ""
	}

	version: string | *"3"
	volumes: [string]: null
	services: [string]: {
		image: string
		build?: {
			context: string
			args: [string]: string
		}
		depends_on?: [...string]
		ports?: [...string]
		environment?: [string]: string
		command?: [...string]
		volumes?: [...string]
		restart: string
	}
}

// add a compose service
#AddComposeService: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	restart:    _
	containers: _
	$metadata:  _
	$dependencies: [...string]
	$resources: compose: _#ComposeResource & {
		services: "\($metadata.id)": {
			image: containers.default.image
			environment: {
				for key, value in containers.default.env {
					"\(key)": {
						if (value & string) != _|_ {
							value
						}
						if (value & v1.#Secret) != _|_ {
							[
								if value.version == _|_ {"SECRET[\(value.name)]"},
								if value.version != _|_ {"SECRET[\(value.name):\(value.version)]"},
							][0]
						}
					}
				}
			}
			depends_on: [
				for id in $dependencies if services[id] != _|_ {id},
			]
			command: list.Concat([
					containers.default.command,
					containers.default.args,
			])
			"restart": [
					if restart == "always" {"always"},
					if restart == "onfail" {"on-failure"},
					if restart == "never" {"no"},
			][0]
			volumes: [
				for m in containers.default.mounts {
					_mapping: [
							if m.volume.local != _|_ {"\(m.volume.local):\(m.path)"},
							if m.volume.persistent != _|_ {"\(m.volume.persistent):\(m.path)"},
							if m.volume.ephemeral != _|_ {"\(m.volume.ephemeral):\(m.path)"},
							if m.volume.secret != _|_ {"\(m.volume.secret.name):\(m.path)"},
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

// add a compose service
#AddComposeVolume: v1.#Transformer & {
	v1.#Component
	traits.#Volume
	volumes: _
	$dependencies: [...string]
	$resources: compose: _#ComposeResource & {
		for k, v in volumes {
			if v.persistent != _|_ {
				volumes: "\(v.persistent)": null
			}

			// support ephemeral for data sharing
			if v.ephemeral != _|_ {
				volumes: "\(v.ephemeral)": null
			}

			if v.secret != _|_ {
				volumes: "\(v.secret.name)": null
			}
		}
	}
}

// expose a compose service ports
#ExposeComposeService: v1.#Transformer & {
	v1.#Component
	traits.#Exposable
	$metadata: _
	$dependencies: [...string]
	endpoints: default: host: "\($metadata.id)"
	$resources: compose: _#ComposeResource & {
		services: "\($metadata.id)": {
			ports: [
				for p in endpoints.default.ports {
					"\(p.port):\(p.target)"
				},
			]
		}
	}
}

// add a compose service for a postgres database
#AddComposePostgres: v1.#Transformer & {
	v1.#Component
	traits.#Postgres
	$dependencies: [...string]
	version:    _
	persistent: _
	port:       _
	database:   _
	$metadata:  _
	host:       "\($metadata.id)"
	username:   string @guku(generate)
	password:   string @guku(generate,secret)
	$resources: compose: _#ComposeResource & {
		services: "\($metadata.id)": {
			image: "postgres:\(version)-alpine"
			ports: [
				"\(port)",
			]
			if persistent {
				volumes: [
					"pg-data:/var/lib/postgresql/data",
				]
			}
			environment: {
				POSTGRES_USER:     username
				POSTGRES_PASSWORD: password
				POSTGRES_DB:       database
			}
			depends_on: [
				for id in $dependencies if services[id] != _|_ {id},
			]
			restart: "no"
		}
		if persistent {
			volumes: "pg-data": null
		}
	}
}

// add compose build filed to build an image locally
#AddComposeBuild: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	build: {
		context: string | *"."
		args: [string]: string
	}
	$metadata: _
	$resources: compose: _#ComposeResource & {
		services: "\($metadata.id)": "build": build
	}
}
