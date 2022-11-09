package traits

import (
	"list"
	"guku.io/devx"
)

#Workload: devx.#Trait & {
	$guku: traits: Workload: null

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

#Replicable: devx.#Trait & {
	$guku: traits: Replicable: null

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
}

#Exposable: devx.#Trait & {
	$guku: traits: Exposable: null

	ports: [...{
		port:   uint
		target: uint | *port
	}] & list.MinItems(0)
	host: string
}

#Postgres: devx.#Trait & {
	$guku: traits: Postgres: null
}

#Mysql: devx.#Trait & {
	$guku: traits: Mysql: null
}
