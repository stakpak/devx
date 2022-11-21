# guku DevX

[![asciicast](https://asciinema.org/a/cIhBiPlYmIok6H5nsJzmGCoDy.svg)](https://asciinema.org/a/cIhBiPlYmIok6H5nsJzmGCoDy)

## Developer self-service with a single config file for all envrionments, for all vendors!

![alt text](https://devx.guku.io/assets/images/image02.png)


## Quick start
```bash
➜ devx project init
➜ devx project update
➜ devx project gen
➜ devx do dev  
```

## Usage

### Validation
```bash
➜ devx project validate
```

### Platform capability discovery
```bash
➜ devx project discover --transformers
[🏷️  traits] "guku.io/devx/v1/traits"
traits.#Workload        a component that runs a container 
traits.#Replicable      a component that can be horizontally scaled 
traits.#Exposable       a component that has endpoints that can be exposed 
traits.#Postgres        a postgres database 

[🏭 transformers] "guku.io/devx/v1/transformers/compose"
compose.#AddComposeService      add a compose service  (requires trait:Workload)
compose.#ExposeComposeService   expose a compose service ports  (requires trait:Exposable)
compose.#AddComposePostgres     add a compose service for a postgres database  (requires trait:Postgres)
```

## Package management

devx can pull CUE code directly from git repositories.

### Create a new packages
Create a new repository to store your packages, you can host multiple packages in the same repository.

```bash
pkg
└── domain.com
    └── package1
        ├── cue.mod
        |   └── module.cue # module: "domain.com/package1"
        └── file.cue
```

### Add the package to `module.cue`
```cue
module: ""

packages: [
	"github.com/<org name>/<repo name>@<revision>/pkg/domain.com",
]		
```

### For private packages (optional)
```bash
export GIT_USERNAME="username"
export GIT_PASSWORD="password"
```
or
```bash
export GIT_PRIVATE_KEY_FILE="path/to/key"
export GIT_PRIVATE_KEY_FILE_PASSWORD="password"

```

### Update packages (pulling updates will replace existing packages)
```
➜ devx project update
```
