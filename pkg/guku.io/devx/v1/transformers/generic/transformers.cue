package generic

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

#AddExtraEnv: v1.#Transformer & {
	$metadata: transformer: "AddExtraEnv"
	args: {
		env: [string]: string
	}
	context: {
		dependencies: [...string]
	}
	input: {
		v1.#Component
		traits.#Workload
		...
	}
	output: {
		containers: [string]: env: {
			for k, v in args.env {
				"\(k)": v
			}
		}
	}
}
