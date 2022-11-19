package database

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

// transform a postgres database definition to a workload
#AddDatabaseWorkload: v1.#Transformer & {
	$metadata: transformer: "AddDatabaseWorkload"

	context: {
		dependencies: [...string]
	}

	_defaults: {
		if input.database.type == "postgres" {
			image: "postgres:\(input.database.version)-alpine"
			port:  "3306"
		}
	}

	ARGS=args: {
		image: string | *_defaults.image
		port:  string | *_defaults.port
	}

	input: {
		v1.#Component
		traits.#Database
		...
	}
	output: {
		v1.#Component
		traits.#Workload
		traits.#Exposable

		containers: default: {
			image: ARGS.image
		}

		endpoints: default: port: ARGS.port
	}
}

#AddDatabaseVolume: v1.#Transformer & {
	$metadata: transformer: "AddDatabaseVolume"

	context: {
		dependencies: [...string]
	}

	args: {}

	input: {
		v1.#Component
		traits.#Database & {database: persistent: true}
		traits.#Workload
		...
	}
	output: {
		v1.#Component
		traits.#Volume

		containers: default: mounts: data: {
			volume: output.volumes.data
			path: "/var/lib/pgsql/data"
		}

		volumes: data: {}
	}
}
