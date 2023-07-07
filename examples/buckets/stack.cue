package main

import (
	"stakpak.dev/devx/v1"
	"stakpak.dev/devx/v1/traits"
)

stack: v1.#Stack
stack: {
	$metadata: stack: "myapp"
	components: {
		bucket: {
			traits.#S3CompatibleBucket
			s3: {
				prefix:        "guku-io-"
				name:          "my-bucket-123"
				versioning:    false
				objectLocking: false
			}
		}
	}
}
