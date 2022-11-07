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
	$children: _
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

	feedforward: {
		components: [string]: _
	}

	feedback: {
		component: input.component
	}
}

// func feedforward(context, component) -> components
// func feedback(components) -> component
