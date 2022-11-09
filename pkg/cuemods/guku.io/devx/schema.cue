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
