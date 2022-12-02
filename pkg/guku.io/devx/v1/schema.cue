package v1

import "list"

#Trait: {
	#Component
	$metadata: traits: [string]: _ | *null
	...
}

#Component: {
	$metadata: {
		id: string
		labels: [string]: string
	}
	#Trait
}

#Stack: {
	$metadata: "Stack"
	components: [Id=string]: #Component & {
		$metadata: id: Id
	}
}

#Transformer: {
	#Component
	$resources: [string]: {
		$metadata: labels: [string]: string
		$metadata: labels: driver:   string
		$metadata: labels: type:     string
		...
	}
}

#StackBuilder: {
	[string]: {
		// we might not use this at all in V1
		additionalComponents?: [Id=string]: #Component & {
			$metadata: id: Id
		}

		preflows: [...#Flow]
		mainflows: [...#Flow]
		postflows: [...#Flow]

		flows: list.Concat([
			preflows,
			mainflows,
			postflows,
		])
	}
}

#Flow: {
	match: {
		traits: [string]: _
		labels: [string]: string
	}
	exclude: {
		traits: [string]: _
		labels: [string]: string
	}

	pipeline: [...#Transformer]

	// include all transformer traits by default in match
	for t in pipeline {
		match: traits: t.$metadata.traits
	}
}
