package devx

#Application: {
	$guku: "Application"

	components: #Components
}

#Components: [Id=string]: {
	#Component
	...
} & {
	$guku: id: Id
}

#Component: {
	$guku: {
		component: string
		id:        string
		traits: [string]: _
		children?: _
	}
}

#Workload: {
	#Component
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

#Transformer: {
	$guku: transformer: {
		name:      string
		component: string
	}

	input: {
		context: _
		component: {
			#Component
			...
		}
	}

	feedforward: {
		components: #Components
	}

	feedback: {
		component: input.component
	}
}

// func feedforward(context, component) -> components
// func feedback(components) -> component
