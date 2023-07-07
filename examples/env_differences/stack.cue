package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		cowsay: {
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
