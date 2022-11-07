package compose

import "guku.io/devx"

#ComposeManifest: {
	version: string | *"3"
	volumes: [string]: null
	services: [string]: {
		image: string
		depends_on?: [...string]
		ports?: [...string]
		environment?: [string]: string
		command?: string
		volumes?: [...string]
	}
}

// #ComposeService: devx.#Transformer & {
//  $guku: transformer: {
//   name:      "ComposeService"
//   component: "Service"
//  }

//  input: {
//   component: devx.#Service
//   context: {
//    dependencies: [...string]
//   }
//  }

//  output: {
//   components: {
//    compose: #ComposeManifest & {
//     services: "\(input.component.id)": {
//      image: input.component.image
//      ports: [
//       for p in input.component.ports {
//        "\(p.port):\(p.target)"
//       },
//      ]
//      environment: input.component.env
//      depends_on:  input.context.dependencies
//     }
//    }
//   }
//   propagate: {
//    host: "\(input.component.id)"
//   }
//  }
// }

#ComposeService: devx.#Transformer & {
	$guku: transformer: {
		name:      "ComposeService"
		component: "Service"
	}

	input: {
		component: devx.#Service
		context: {
			dependencies: [...string]
		}
	}

	feedforward: components: compose: #ComposeManifest & {
		services: "\(input.component.id)": {
			image: input.component.image
			ports: [
				for p in input.component.ports {
					"\(p.port):\(p.target)"
				},
			]
			environment: input.component.env
			depends_on:  input.context.dependencies
		}
	}

	feedback: component: {
		host: "\(input.component.id)"
	}

}

// #ComposeWorker: devx.#Transformer & {
//  $guku: transformer: {
//   name:      "ComposeWorker"
//   component: "Worker"
//  }

//  input: {
//   component: devx.#Worker
//   context: {
//    dependencies: [...string]
//   }
//  }

//  output: {
//   components: {
//    compose: #ComposeManifest & {
//     services: "\(input.component.id)": {
//      image:       input.component.image
//      environment: input.component.env
//      depends_on:  input.context.dependencies
//     }
//    }
//   }
//   propagate: {}
//  }
// }

// #ComposeJob: devx.#Transformer & {
//  $guku: transformer: {
//   name:      "ComposeJob"
//   component: "Job"
//  }

//  input: {
//   component: devx.#Job
//   context: {
//    dependencies: [...string]
//   }
//  }

//  output: {
//   components: {
//    compose: #ComposeManifest & {
//     services: "\(input.component.id)": {
//      image:       input.component.image
//      environment: input.component.env
//      depends_on:  input.context.dependencies
//     }
//    }
//   }
//   propagate: {}
//  }
// }

// #ComposeCronJob: devx.#Transformer & {
//  $guku: transformer: {
//   name:      "ComposeCronJob"
//   component: "CronJob"
//  }

//  input: {
//   component: devx.#CronJob
//   context: {
//    dependencies: [...string]
//   }
//  }

//  output: {
//   components: {
//    compose: #ComposeManifest & {
//     services: "\(input.component.id)": {
//      image:       input.component.image
//      environment: input.component.env
//      depends_on:  input.context.dependencies
//     }
//    }
//   }
//   propagate: {}
//  }
// }

// #ComposePostgresDB: devx.#Transformer & {
//  $guku: transformer: {
//   name:      "ComposePostgresDB"
//   component: "PostgresDB"
//  }

//  input: {
//   component: devx.#PostgresDB
//   context: {
//    dependencies: [...string]
//   }
//  }

//  output: {
//   components: {
//    _username: string @guku(generate)
//    _password: string @guku(generate,secret)
//    compose:   #ComposeManifest & {
//     services: "\(input.component.id)": {
//      image: "postgres:\(input.component.version)-alpine"
//      ports: [
//       "\(input.component.port)",
//      ]
//      if input.component.persistent {
//       volumes: [
//        "pg-data:/var/lib/postgresql/data",
//       ]
//      }
//      environment: {
//       POSTGRES_USER:     _username
//       POSTGRES_PASSWORD: _password
//       POSTGRES_DB:       input.component.database
//      }
//      depends_on: input.context.dependencies
//     }
//     if input.component.persistent {
//      volumes: "pg-data": null
//     }
//    }
//   }
//   propagate: {
//    host:     "\(input.component.id)"
//    username: components._username
//    password: components._password
//   }
//  }
// }

#ComposePostgresDB: devx.#Transformer & {
	$guku: transformer: {
		name:      "ComposePostgresDB"
		component: "PostgresDB"
	}

	input: {
		component: devx.#PostgresDB
		context: {
			dependencies: [...string]
		}
	}

	feedforward: components: {
		_username: string @guku(generate)
		_password: string @guku(generate,secret)
		compose:   #ComposeManifest & {
			services: "\(input.component.id)": {
				image: "postgres:\(input.component.version)-alpine"
				ports: [
					"\(input.component.port)",
				]
				if input.component.persistent {
					volumes: [
						"pg-data:/var/lib/postgresql/data",
					]
				}
				environment: {
					POSTGRES_USER:     _username
					POSTGRES_PASSWORD: _password
					POSTGRES_DB:       input.component.database
				}
				depends_on: input.context.dependencies
			}
			if input.component.persistent {
				volumes: "pg-data": null
			}
		}
	}

	feedback: component: {
		host:     "\(input.component.id)"
		username: feedforward.components._username
		password: feedforward.components._password
	}
}
