package traits

import (
	"struct"
	
	"guku.io/devx/v1"
)

#Workload: v1.#Trait & {
	$metadata: traits: Workload: null

	containers: [string]: {
		image: string
		command: [...string]
		args: [...string]
		env: [string]: string
		mounts: [string]: {
			volume: {...}
			path: string
		}
	}
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

	endpoints: [string]: {
		host: string
		port: string
	} & struct.MinFields(0)
}

#Volume: v1.#Trait & {
	v1.#Trait
	$metadata: traits: Volume: null

	volumes: [Id=string]: {
		$metadata: id: Id
		...
	}
}

#Database: v1.#Trait & {
	$metadata: traits: Database: null

	database: {
		type:       string
		version:    string | *"latest"
		persistent: bool | *true
		database:   string | *"default"
		username?:  string
		password?:  string
	}
}
