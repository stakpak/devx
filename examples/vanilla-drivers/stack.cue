package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v2alpha1"
)

stack: v1.#Stack & {
	components: {
		app1: yo: 1
		app2: yo: 1
	}
}

#YAMLResource: {
	$metadata: labels: {
		driver: "yaml"
		type:   ""
	}
	...
}
#JSONResource: {
	$metadata: labels: {
		driver: "json"
		type:   ""
	}
	...
}
builders: v2alpha1.#Environments & {
	config: v2alpha1.#StackBuilder & {
		drivers: {
			yaml: output: {
				dir: ["."]
				file: "conf.yml"
			}
			json: output: {
				dir: ["."]
				file: "conf.json"
			}
		}
		flows: "yaml": pipeline: [{
			$metadata: _
			yo:        _
			$resources: {
				config: #YAMLResource & {
					"\($metadata.id)": yo
				}
				other: #JSONResource & {
					"\($metadata.id)": yo
				}
			}
		}]
	}
}
