package devx

#Workload: {
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

#Replicable: {
	$guku: traits: Replicable: null

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
}
