package traits

import (
	"list"
	"guku.io/devx/v1"
)

#Workload: v1.#Trait & {
	$metadata: traits: Workload: null

	image: string
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

#Replicable: v1.#Trait & {
	$metadata: traits: Replicable: null

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
}

#Exposable: v1.#Trait & {
	$metadata: traits: Exposable: null

	ports: [...{
		port:   uint
		target: uint | *port
	}] & list.MinItems(0)
	host: string
}

#Postgres: v1.#Trait & {
	$metadata: traits: Postgres: null

	version:    string
	persistent: bool | *true
	port:       uint | *5432
	database:   string | *"default"

	host:     string
	username: string
	password: string
	url:      "postgresql://\(username):\(password)@\(host):\(port)/\(database)"
}
