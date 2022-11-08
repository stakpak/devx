package devx

#Application: {
	$guku: "application"

	components: #Components
}

#Components: [Id=string]: {
	#Component
	...
} & {
	$id: Id
}

#Component: {
	$guku: component: string

	$id:        string
	$children?: _
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
