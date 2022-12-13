package aws

import (
	"strings"
	"guku.io/devx/v1"
	"guku.io/devx/v1/traits"
)

_#TerraformResource: {
	$metadata: labels: {
		driver: "terraform"
		type:   ""
	}
}

// add a parameter store secret
#AddSSMSecretParameter: v1.#Transformer & {
	v1.#Component
	traits.#Secret
	$metadata: _
	secrets:   _
	$resources: terraform: {
		_#TerraformResource
		resource: {
			aws_ssm_parameter: {
				for key, secret in secrets {
					"\(strings.ToLower($metadata.id))_\(strings.ToLower(key))": {
						name:  secret.key
						type:  "SecureString"
						value: "${random_password.\(strings.ToLower($metadata.id))_\(strings.ToLower(key)).result}"
					}
				}
			}
			random_password: {
				for key, _ in secrets {
					"\(strings.ToLower($metadata.id))_\(strings.ToLower(key))": {
						length:  32
						special: false
					}
				}
			}
		}
	}
}
