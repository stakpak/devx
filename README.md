# guku DevX

## Quick start
```bash
devx project init
devx project update
devx project gen
devx do dev  
```

## Usage

### Validation
```bash
devx project validate
```

### Platform capability discovery
```bash
devx project discover
devx project discover -s # to show schemas
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
devx project update
```