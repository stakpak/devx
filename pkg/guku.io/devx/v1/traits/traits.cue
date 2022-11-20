package traits

import (
	"list"
	"guku.io/devx/v1"
)

_#Container: {
	image: string @guku(required)
	command: [...string]
	args: [...string]
	env: [string]:    string
	mounts: [string]: string
	volumes: [...{
		source:   string
		target:   string
		readOnly: bool | *true
	}]
}

// a component that runs a container
#Workload: v1.#Trait & {
	$metadata: traits: Workload: null

	containers: [string]: _#Container
	containers: default:  _#Container
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
