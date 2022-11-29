package traits

import (
	"list"
	"guku.io/devx/v1"
)

_#Container: {
	image: string @guku(required)
	command: [...string]
	args: [...string]
	env: [string]: string
	mounts: [...{
		volume:   _#VolumeSpec
		path:     string
		readOnly: bool | *true
	}]
	resources: {
		requests?: {
			cpu?:    string
			memory?: string
		}
		limits?: {
			cpu?:    string
			memory?: string
		}
	}
}

// a component that runs containers
#Workload: v1.#Trait & {
	$metadata: traits: Workload: null

	containers: [string]: _#Container
	containers: default:  _#Container
	restart: "onfail" | "never" | *"always"
}

// a component that can be horizontally scaled
#Replicable: v1.#Trait & {
	$metadata: traits: Replicable: null

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
}

_#Endpoint: {
	ports: [...{
		port:   uint
		target: uint | *port
	}] & list.MinItems(0)
	host: string
}

// a component that has endpoints that can be exposed
#Exposable: v1.#Trait & {
	$metadata: traits: Exposable: null

	endpoints: [string]: _#Endpoint
	endpoints: default:  _#Endpoint
}

_#VolumeSpec: {
	local: string
} | {
	ephemeral: string
} | {
	persistent: string
}

// a component that has a volume
#Volume: v1.#Trait & {
	$metadata: traits: Volume: null

	volumes: [string]: _#VolumeSpec
	volumes: default:  _#VolumeSpec
}

// a postgres database
#Postgres: v1.#Trait & {
	$metadata: traits: Postgres: null

	version:    string @guku(required)
	persistent: bool | *true
	port:       uint | *5432
	database:   string | *"default"

	host:     string
	username: string
	password: string
	url:      "postgresql://\(username):\(password)@\(host):\(port)/\(database)"
}

_#HelmCommon: {
	chart:     string @guku(required)
	url:       string @guku(required)
	version:   string @guku(required)
	values:    _ | *{}
	namespace: string
}

// a helm chart using helm repo
#Helm: v1.#Trait & {
	$metadata: traits: Helm: null
	_#HelmCommon
}

// a helm chart using git
#HelmGit: v1.#Trait & {
	$metadata: traits: HelmGit: null
	_#HelmCommon
}

// a helm chart using oci
#HelmOCI: v1.#Trait & {
	$metadata: traits: HelmOCI: null
	_#HelmCommon
}
