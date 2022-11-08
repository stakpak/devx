package devx

import "list"

#Service: {
	#Workload
	$guku: component: "Service"

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
	ports: [...{
		port:   uint
		target: uint | *port
	}] & list.MinItems(0)
	host: string
}

#Worker: {
	#Workload
	$guku: component: "Worker"

	replicas: {
		min: uint | *1
		max: uint & >=min | *min
	}
}

#Job: {
	#Workload
	$guku: component: "Job"

	backoffLimit: uint | *0
}

#CronJob: {
	#Workload
	$guku: component: "CronJob"

	schedule: string
}

#PostgresDB: {
	#Component
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
	#Component
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
