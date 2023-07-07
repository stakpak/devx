package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		sharedvol: {
			traits.#Volume
			volumes: default: persistent: "bazo"
		}
		cowsay: {
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
