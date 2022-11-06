package devx

#Application: {
	$guku: "application"

	components: [Id=string]: {
		#Component
		...
	} & {
		id: Id
	}
}

#Component: {
	$guku: component: string

	id:        string
	outputs:   _
	$manifest: _
}

#Workload: {
	#Component

	image: string
	command: [...string]
	args: [...string]
	env: [string]: string
}

#Transformer: {
	$guku: transformer: {
		name:      string
		component: string
	}

	input: {
		context:   _
		component: _
	}

	output: {
		components: _
		propagate:  _
	}
}
