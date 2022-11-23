package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		cowsay: {
			v1.#Component
			traits.#Workload
			containers: default: {
				image: "docker/whalesay"
				command: ["cowsay"]
			}
		}
	}
}

builders: {
	dev: additionalComponents: cowsay: containers: default: args: ["Hello DEV!"]
	prod: additionalComponents: cowsay: containers: default: args: ["Hello Prod!"]
}
