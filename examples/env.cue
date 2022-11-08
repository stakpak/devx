package main

import (
	"guku.io/devx/transformers/compose"
)

environments: {
	dev: {
		Service:    compose.#ComposeService
		PostgresDB: compose.#ComposePostgresDB
		MysqlDB:    compose.#ComposeMysqlDB
	}
}
