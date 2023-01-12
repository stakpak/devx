package main

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/components"
	"guku.io/devx/v1/transformers/compose"
	tfaws "guku.io/devx/v1/transformers/terraform/aws"
)

builders: v1.#StackBuilder

builders: prod: mainflows: [
	{
		pipeline: [tfaws.#AddS3Bucket]
	},
]

builders: dev: mainflows: [
	{
		pipeline: [compose.#AddComposeService]
	},
	{
		pipeline: [compose.#ExposeComposeService]
	},
	{
		pipeline: [compose.#AddComposeVolume]
	},
	{
		pipeline: [compose.#AddS3Bucket]
	},
]

builders: dev: additionalComponents: {
	myminio: {
		components.#Minio
		minio: {
			urlScheme: "http"
			userKeys: default: {
				accessKey:    "admin"
				accessSecret: "adminadmin"
			}
			url: _
		}
	}
	bucket: s3: {
		url:          myminio.minio.url
		accessKey:    myminio.minio.userKeys.default.accessKey
		accessSecret: myminio.minio.userKeys.default.accessSecret
	}
}
