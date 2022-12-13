package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		commonSecrets: {
			traits.#Secret
			secrets: apiKey: {
				name:    "apikey-a"
				version: "4"
			}
		}
		cowsay: {
			traits.#Workload
			traits.#Volume
			containers: default: {
				image: "docker/whalesay"
				command: ["cowsay"]
				args: ["Hello DevX!"]
				env: {
					API_KEY:   commonSecrets.secrets.apiKey
					SOMETHING: "bla"
				}
				mounts: [
					{
						volume: volumes.default
						path:   "secrets/file"
					},
				]
			}
			volumes: default: secret: commonSecrets.secrets.apiKey
		}
	}
}
