## [Documentation](https://devx.stakpak.dev/docs/intro)

## Introduction

DevX is a tool for building lightweight Internal Developer Platforms. Use DevX to build internal standards, prevent misconfigurations early, and enable infrastructure self-service.

## Installation

### Homebrew
```bash
brew tap stakpak/stakpak
brew install devx       
```

### Download the binary

[Releases page](https://github.com/stakpak/devx/releases)

### Docker image
```bash
docker run --rm -v "$(pwd):/app" ghcr.io/stakpak/devx:latest -h
```

## Quick start
```bash
➜ devx project init
➜ devx project update
➜ devx project gen
➜ devx build dev
🏭 Transforming stack for the "dev" environment...
[compose] applied resources to "build/dev/compose/docker-compose.yml"
[terraform] applied resources to "build/dev/terraform/generated.tf.json"
```

![demo](assets/demo.gif)


## Usage

### Configuration language
We use [CUE](https://cuelang.org/) to write strongly typed configurations. You can now shift YAML typos left, instead of detecting errors when applying configurations. You can easily transform CUE configuration files to and from YAML (CUE is a superset of YAML & JSON).

[CUE](https://cuelang.org/) is the result of years of experience writing configuration languages at Google, and seeks to improve the developer experience while avoiding some nasty pitfalls. CUE looks like JSON, while making declarative data definition, generation, and validation a breeze. You can find a primer on CUE [here](https://docs.dagger.io/1215/what-is-cue/#understanding-cue).


### Create a stack (by Developers)
You create a stack to define the workload and its dependencies.
```cue
package main

import (
    "stakpak.dev/devx/v1"
    "stakpak.dev/devx/v1/traits"
)

stack: v1.#Stack & {
    components: {
        cowsay: {
            traits.#Workload
            containers: default: {
                image: "docker/whalesay"
                command: ["cowsay"]
                args: ["Hello DevX!"]
            }
        }
    }
}
```

### Create your own stack builders or use community packages (by Platform Engineers)
You can customize how the stack is processed by writing declarative transformers.
```cue
package main

import (
    "stakpak.dev/devx/v2alpha1"
    "stakpak.dev/devx/v2alpha1/environments"
)

builders: v2alpha1.#Environments & {
    dev: environments.#Compose
}
```

### Validation
Validate configurations while writing
```bash
➜ devx project validate
👌 Looks good
```

### Platform capability discovery
```bash
➜ devx project discover --transformers
[🏷️  traits] "stakpak.dev/devx/v1/traits"
traits.#Workload	a component that runs a container 
traits.#Replicable	a component that can be horizontally scaled 
traits.#Exposable	a component that has endpoints that can be exposed 
traits.#Postgres	a postgres database 
traits.#Helm	a helm chart using helm repo 
traits.#HelmGit	a helm chart using git 
traits.#HelmOCI	a helm chart using oci 

[🏭 transformers] "stakpak.dev/devx/v1/transformers/argocd"
argocd.#AddHelmRelease	add a helm release  (requires trait:Helm)

[🏭 transformers] "stakpak.dev/devx/v1/transformers/compose"
compose.#AddComposeService	add a compose service  (requires trait:Workload)
compose.#ExposeComposeService	expose a compose service ports  (requires trait:Exposable)
compose.#AddComposePostgres	add a compose service for a postgres database  (requires trait:Postgres)

[🏭 transformers] "stakpak.dev/devx/v1/transformers/terraform"
terraform.#AddHelmRelease	add a helm release  (requires trait:Helm)
```

## Package management

You can publish and share CUE packages directly through git repositories.

### Create a new packages
Create a new repository to store your packages (you can host multiple packages per repository).

```bash
cue.mod
└── module.cue # module: "domain.com/platform"
subpackage
└── file.cue
file.cue
```

### Add the package to `module.cue`
```cue
module: ""

packages: [
  "github.com/<org name>/<repo name>@<git revision>:",
]       	
```

### For private packages (optional)
```bash
export GIT_USERNAME="username"
export GIT_PASSWORD="password"
```

### Update packages (pulling updates will replace existing packages)
```
➜ devx project update
```

## Contributors

<table>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/kajogo777>
            <img src=https://avatars.githubusercontent.com/u/10531031?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=George/>
            <br />
            <sub style="font-size:14px"><b>George</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/thethumbler>
            <img src=https://avatars.githubusercontent.com/u/3092919?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Mohamed Hamza/>
            <br />
            <sub style="font-size:14px"><b>Mohamed Hamza</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/tranngoclam>
            <img src=https://avatars.githubusercontent.com/u/4991619?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Lam Tran/>
            <br />
            <sub style="font-size:14px"><b>Lam Tran</b></sub>
        </a>
    </td>
        <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/ahmedhesham6>
            <img src=https://avatars.githubusercontent.com/u/23265119?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ahmed Hesham/>
            <br />
            <sub style="font-size:14px"><b>Ahmed Hesham</b></sub>
        </a>
    </td>
</tr>
</table>
