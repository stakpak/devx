package components

import (
	"guku.io/devx"
	"guku.io/devx/traits"
)

import "list"

#Service: {
	devx.#Component
	traits.#Workload
	traits.#Replicable
	$guku: component: "Service"

	ports: [...{
		port:   uint
		target: uint | *port
	}] & list.MinItems(0)
	host: string
}

#Worker: {
	devx.#Component
	traits.#Workload
	traits.#Replicable
	$guku: component: "Worker"
}

#Job: {
	devx.#Component
	traits.#Workload
	$guku: component: "Job"

	backoffLimit: uint | *0
}

#CronJob: {
	devx.#Component
	traits.#Workload
	$guku: component: "CronJob"

	schedule: string
}

#PostgresDB: {
	devx.#Component
	$guku: component: "PostgresDB"

	version:    string
	persistent: bool | *true
	port:       uint | *5432
	database:   string | *"default"

	host:     string
	username: string
	password: string
	url:      "postgresql://\(username):\(password)@\(host):\(port)/\(database)"
}

#MysqlDB: {
	devx.#Component
	$guku: component: "MysqlDB"

	version:    string
	persistent: bool | *true
	port:       uint | *3306
	database:   string | *"default"

	host:     string
	username: string
	password: string
	url:      "mysql://\(username):\(password)@\(host):\(port)/\(database)"
}
