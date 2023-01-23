package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

stack: v1.#Stack & {
	components: {
		cowsay: {
			traits.#Helm
			helm: {
				k8s: version: minor: 19
				url:       "guku.io"
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
