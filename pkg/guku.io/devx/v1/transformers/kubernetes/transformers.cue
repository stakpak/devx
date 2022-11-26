package kubernetes

import (
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

_#DeploymentResource: {
	appsv1.#Deployment
	$metadata: labels: driver: "kubernetes"
	kind:       "Deployment"
	apiVersion: "apps/v1"
}
_#ServiceAccountResource: {
	corev1.#ServiceAccount
	$metadata: labels: driver: "kubernetes"
	kind:       "ServiceAccount"
	apiVersion: "v1"
}
_#ServiceResource: {
	corev1.#Service
	$metadata: labels: driver: "kubernetes"
	kind:       "Service"
	apiVersion: "v1"
}

#AddDeployment: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata:  _
	containers: _

	appName:            string | *$metadata.id
	serviceAccountName: string | *$metadata.id
	$resources: {
		"\(appName)-deployment": _#DeploymentResource & {
			metadata: {
				name: appName
				labels: app: appName
			}
			spec: {
				selector: matchLabels: app: appName
				template: {
					metadata: {
						annotations: {}
						labels: app: appName
					}
					spec: {
						"serviceAccountName": serviceAccountName
						securityContext: {
							runAsUser:  1099
							runAsGroup: 1099
							fsGroup:    1099
						}
						"containers": [
							for k, container in containers {
								{
									name:    k
									image:   container.image
									command: container.command
									args:    container.args
									env: [
										for name, value in container.env {
											{
												"name":  name
												"value": value
											}
										},
									]
									if container.resources.limits.cpu != _|_ {
										resources: limits: cpu: container.resources.limits.cpu
									}
									if container.resources.limits.memory != _|_ {
										resources: limits: memory: container.resources.limits.memory
									}
									if container.resources.requests.cpu != _|_ {
										resources: requests: cpu: container.resources.requests.cpu
									}
									if container.resources.requests.memory != _|_ {
										resources: requests: memory: container.resources.requests.memory
									}
								}
							},
						]
					}
				}
			}
		}

		"\(appName)-sa": _#ServiceAccountResource & {
			metadata: {
				name: serviceAccountName
				labels: app: appName
			}
		}
	}
}

#AddService: v1.#Transformer & {
	v1.#Component
	traits.#Exposable
	$metadata: _
	endpoints: _

	appName: string | *$metadata.id
	endpoints: default: host: appName
	$resources: "\(appName)-svc": _#ServiceResource & {
		metadata: name: "\(appName)"
		spec: {
			selector: app: "\(appName)"
			ports: [
				for p in endpoints.default.ports {
					{
						name: "\(p.port)"
						port: p.port
					}
				},
			]
		}
	}
}

#AddReplicas: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	traits.#Replicable
	$metadata: _
	replicas:  _

	appName: string | *$metadata.id
	$resources: {
		"\(appName)-deployment": _#DeploymentResource & {
			spec: "replicas": replicas.min
		}
	}
}

#AddPodLabels: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _
	podLabels: [string]: string

	appName: string | *$metadata.id
	$resources: {
		"\(appName)-deployment": _#DeploymentResource & {
			spec: template: metadata: labels: podLabels
		}
	}
}

#AddPodAnnotations: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _
	podAnnotations: [string]: string

	appName: string | *$metadata.id
	$resources: {
		"\(appName)-deployment": _#DeploymentResource & {
			spec: template: metadata: annotations: podAnnotations
		}
	}
}
