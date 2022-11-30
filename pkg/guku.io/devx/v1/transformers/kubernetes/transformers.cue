package kubernetes

import (
	"strings"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

_#KubernetesMeta: {
	metadata?: metav1.#ObjectMeta
	...
}
_#WorkloadResource: {
	_#KubernetesMeta
	$metadata: labels: {
		driver: "kubernetes"
		type:   strings.HasPrefix("k8s.io/")
	}
	spec: {
		template: corev1.#PodTemplateSpec
		...
	}
}
_#DeploymentResource: {
	appsv1.#Deployment
	$metadata: labels: {
		driver: "kubernetes"
		type:   "k8s.io/apps/v1/deployment"
	}
	kind:       "Deployment"
	apiVersion: "apps/v1"
	spec: template: spec: securityContext: {
		runAsUser:  uint | *10000
		runAsGroup: uint | *10000
		fsGroup:    uint | *10000
	}
}
_#ServiceAccountResource: {
	corev1.#ServiceAccount
	$metadata: labels: {
		driver: "kubernetes"
		type:   "k8s.io/core/v1/serviceaccount"
	}
	kind:       "ServiceAccount"
	apiVersion: "v1"
}
_#ServiceResource: {
	corev1.#Service
	$metadata: labels: {
		driver: "kubernetes"
		type:   "k8s.io/core/v1/service"
	}
	kind:       "Service"
	apiVersion: "v1"
}

#AddDeployment: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata:  _
	restart:    _
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
						restartPolicy:        [
									if restart == "always" {"Always"},
									if restart == "onfail" {"OnFailure"},
									if restart == "never" {"Never"},
						][0]
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
			type: string | *"ClusterIP"
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
	$resources: "\(appName)-deployment": _#DeploymentResource & {
		spec: "replicas": replicas.min
	}
}

#AddPodLabels: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _
	podLabels: [string]: string

	appName: string | *$metadata.id
	$resources: "\(appName)-deployment": _#DeploymentResource & {
		spec: template: metadata: labels: podLabels
	}
}

#AddPodAnnotations: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _
	podAnnotations: [string]: string

	appName: string | *$metadata.id
	$resources: "\(appName)-deployment": _#DeploymentResource & {
		spec: template: metadata: annotations: podAnnotations
	}
}

#AddNamespace: v1.#Transformer & {
	v1.#Component
	namespace: string

	$resources: [_]: {
		$metadata: _
		if $metadata.labels.driver == "kubernetes" {
			_#KubernetesMeta
			metadata: "namespace": namespace
		}
	}
}

#AddLabels: v1.#Transformer & {
	v1.#Component
	labels: [string]: string

	$resources: [_]: {
		$metadata: _
		if $metadata.labels.driver == "kubernetes" {
			_#KubernetesMeta
			metadata: "labels": labels
		}
	}
}

#AddPodTolerations: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _

	podTolerations: [...corev1.#Toleration]

	appName: string | *$metadata.id
	$resources: "\(appName)-deployment": _#DeploymentResource & {
		spec: template: spec: tolerations: podTolerations
	}
}

#AddPodSecurityContext: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	$metadata: _

	podSecurityContext: corev1.#PodSecurityContext

	appName: string | *$metadata.id
	$resources: "\(appName)-deployment": _#DeploymentResource & {
		spec: template: spec: securityContext: podSecurityContext
	}
}

#AddWorkloadVolumes: v1.#Transformer & {
	v1.#Component
	traits.#Workload
	traits.#Volume

	volumes:    _
	containers: _

	$resources: [_]: {
		$metadata: _
		if $metadata.labels.type == "k8s.io/apps/v1/deployment" || $metadata.labels.type == "k8s.io/apps/v1/statefulset" {
			_#WorkloadResource & {
				spec: template: spec: {
					"volumes": [
						for _, volume in volumes {
							// we support only ephemetal volumes for the time being
							if volume.ephemeral != _|_ {
								{
									name: volume.ephemeral
									emptyDir: {}
								}
							}
						},
					]
					"containers": [
						for _, container in containers {
							volumeMounts: [
								for mount in container.mounts {
									{
										name:      mount.volume.ephemeral
										mountPath: mount.path
										readOnly:  mount.readOnly
									}
								},
							]
						},
					]
				}
			}
		}
	}
}
