package v1

import "list"

#Trait: {
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
	$metadata: transformer: string

	args: _

	context: _

	input: #Component

	output: {
		input
		$resources: [string]: {
			$metadata: labels: [string]: string
			$metadata: labels: driver:   string
			...
		}
	}
}

#StackBuilder: {
	[string]: {
		// we might not use this at all in V1
		additionalComponents?: [Id=string]: #Component & {
			$metadata: id: Id
		}

		preFlows: [...#Flow]
		mainFlows: [...#Flow]
		postFlows: [...#Flow]

		flows: list.Concat([
			preFlows,
			mainFlows,
			postFlows,
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
		match: traits: t.input.$metadata.traits
	}
}
