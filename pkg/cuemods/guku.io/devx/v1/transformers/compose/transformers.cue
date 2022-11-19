package compose

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#ComposeResource: {
	$metadata: labels: driver: "compose"

	version: string | *"3"
	volumes: [string]: {...} | *null
	services: [string]: {
		image: string
		depends_on?: [...string]
		ports?: [...string]
		environment?: [string]: string
		command?: string
		volumes?: [...string]
	}
}

#ComposeExposeService: v1.#Transformer & {
	$metadata: transformer: "ComposeExposeService"

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
		endpoints: [string]: host: "\(input.$metadata.id)"
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
		...
	}
	output: {
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				image: input.containers.default.image
				environment: input.containers.default.env
				depends_on:  context.dependencies
				volumes: [
					for id, mount in output.containers.default.mounts {
						"\(input.$metadata.id)-\(mount.volume.$metadata.id):\(mount.path)"
					},
				]
			}
		}
	}
}

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
			volumes: {
				for id, volume in input.volumes {
					"\(input.$metadata.id)-\(id)": {

					}
				}
			}
		}
	}
}

#ForwardPort: v1.#Transformer & {
	$metadata: transformer: "ForwardPort"

	args: {
		endpoints: [string]: {
			port: string
		}
	}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Exposable
		...
	}
	output: {
		$resources: compose: _#ComposeResource & {
			services: "\(input.$metadata.id)": {
				ports: [
					for id, endpoint in args.endpoints {
						"\(endpoint.port):\(input.endpoints[id].port)"
					}
				]
			}
		}
	}
}