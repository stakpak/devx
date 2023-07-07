package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		cowsay: {
			traits.#Helm
			helm: {
				k8s: version: minor: 19
				url:       "stakpak.dev"
				chart:     "guku"
				version:   "v1"
				namespace: "somethingelse"
				values: {
					bla: 123
				}
			}
		}
	}
}
