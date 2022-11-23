package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		sharedvol: {
			v1.#Component
			traits.#Volume
			volumes: default: persistent: "bazo"
		}
		cowsay: {
			v1.#Component
			traits.#Workload
			containers: default: {
				image: "docker/whalesay"
				command: ["cowsay"]
				args: ["Hello DevX!"]
				mounts: [
					{
						volume:   sharedvol.volumes.default
						path:     "/data/dir"
						readOnly: false
					},
				]
			}
		}
		cowsayagain: {
			v1.#Component
			traits.#Workload
			containers: default: {
				image: "docker/whalesay"
				command: ["cowsay"]
				args: ["Hello again!"]
				mounts: [
					{
						volume:   sharedvol.volumes.default
						path:     "/data/dir"
						readOnly: true
					},
				]
			}
		}
	}
}
