package istio

import (
	"guku.io/devx/v1"
	istiov1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

_#AuthorizationPolicyResource: {
	$metadata: labels: {
		driver: "kubernetes"
		type:   "security.istio.io/v1beta1/authorizationpolicy"
	}
	istiov1beta1.#AuthorizationPolicy
	apiVersion: "security.istio.io/v1beta1"
	kind:       "AuthorizationPolicy"
	status: {}
}

#AddAuthorizationPolicy: v1.#Transformer & {
	v1.#Component
	istioAuthzPolicyName: string
	istioAuthzPolicy: {
		selector: _
		action:   string
		rules:    _
	}

	$resources: "\(istioAuthzPolicyName)-authzpolicy-istio": _#AuthorizationPolicyResource & {
		metadata: name: istioAuthzPolicyName
		spec: istioAuthzPolicy
	}
}
