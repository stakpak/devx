package gitlab

import (
	"list"
	"strings"
)

_#GitlabCISpec: {
	@jsonschema(schema="http://json-schema.org/draft-07/schema#")
	@jsonschema(id="https://gitlab.com/.gitlab-ci.yml")
	$schema?:       string
	image?:         #image
	services?:      #services
	before_script?: #before_script
	after_script?:  #after_script
	variables?:     #globalVariables
	cache?:         #cache
	"!reference"?:  #["!reference"]
	default?: {
		after_script?:  #after_script
		artifacts?:     #artifacts
		before_script?: #before_script
		cache?:         #cache
		image?:         #image
		interruptible?: #interruptible
		retry?:         #retry
		services?:      #services
		tags?:          #tags
		timeout?:       #timeout
		"!reference"?:  #["!reference"]
	}
	stages?:  list.UniqueItems() & [...string] & [_, ...] | *["build", "test", "deploy"]
	include?: #include_item | [...#include_item]
	pages?:   #job
	workflow?: {
		name?: #workflowName
		rules?: [...({
			...
		} | [...string]) & ([...] | {
			if?:        #if
			changes?:   #changes
			exists?:    #exists
			variables?: #rulesVariables
			when?:      "always" | "never"
		})]
		...
	}

	{[=~"^[.]" & !~"^($schema|image|services|before_script|after_script|variables|cache|!reference|default|stages|include|pages|workflow)$"]: (#job_template | _) & _}
	{[!~"^[.]" & !~"^($schema|image|services|before_script|after_script|variables|cache|!reference|default|stages|include|pages|workflow)$"]: #job}

	#artifacts: {
		paths?:     [...string] & [_, ...]
		exclude?:   [...string] & [_, ...]
		expose_as?: string
		name?:      string
		untracked?: bool | *false
		when?:      "on_success" | "on_failure" | "always" | *"on_success"
		expire_in?: string | *"30 days"
		reports?: {
			// Path for file(s) that should be parsed as JUnit XML result
			junit?: string | [...string] & [_, ...]

			// Used to collect coverage reports from the job.
			coverage_report?: {
				// Code coverage format used by the test framework.
				coverage_format?: "cobertura"

				// Path to the coverage report file that should be parsed.
				path?: strings.MinRunes(1)
				...
			}

			// Path to file or list of files with code quality report(s) (such
			// as Code Climate).
			codequality?: #string_file_list

			// Path to file or list of files containing runtime-created
			// variables for this job.
			dotenv?: #string_file_list

			// Path to file or list of files containing code intelligence
			// (Language Server Index Format).
			lsif?: #string_file_list

			// Path to file or list of files with SAST vulnerabilities
			// report(s).
			sast?: #string_file_list

			// Path to file or list of files with Dependency scanning
			// vulnerabilities report(s).
			dependency_scanning?: #string_file_list

			// Path to file or list of files with Container scanning
			// vulnerabilities report(s).
			container_scanning?: #string_file_list

			// Path to file or list of files with DAST vulnerabilities
			// report(s).
			dast?: #string_file_list

			// Deprecated in 12.8: Path to file or list of files with license
			// report(s).
			license_management?: #string_file_list

			// Path to file or list of files with license report(s).
			license_scanning?: #string_file_list

			// Path to file or list of files with performance metrics
			// report(s).
			performance?: #string_file_list

			// Path to file or list of files with requirements report(s).
			requirements?: #string_file_list

			// Path to file or list of files with secret detection report(s).
			secret_detection?: #string_file_list

			// Path to file or list of files with custom metrics report(s).
			metrics?: #string_file_list

			// Path to file or list of files with terraform plan(s).
			terraform?: #string_file_list
			cyclonedx?: #string_file_list
		}
	}

	#string_file_list: string | [...string]

	#include_item: =~"^(https?://|/).+\\.ya?ml$" | {
		// Relative path from local repository root (`/`) to the
		// `yaml`/`yml` file template. The file must be on the same
		// branch, and does not work across git submodules.
		local:  =~"\\.ya?ml$"
		rules?: #rules
	} | {
		// Path to the project, e.g. `group/project`, or
		// `group/sub-group/project` [Learn
		// more](https://docs.gitlab.com/ee/ci/yaml/index.html#includefile).
		project: =~"(?:\\S/\\S|\\$\\S+)"

		// Branch/Tag/Commit-hash for the target project.
		ref?: string
		file: =~"\\.ya?ml$" | [...=~"\\.ya?ml$"]
	} | {
		// Use a `.gitlab-ci.yml` template as a base, e.g.
		// `Nodejs.gitlab-ci.yml`.
		template: =~"\\.ya?ml$"
	} | {
		// URL to a `yaml`/`yml` template file using HTTP/HTTPS.
		remote: =~"^https?://.+\\.ya?ml$"
	}

	#: "!reference": [...strings.MinRunes(1)]

	#image: strings.MinRunes(1) | {
		// Full name of the image that should be used. It should contain
		// the Registry part if needed.
		name: strings.MinRunes(1)

		// Command or script that should be executed as the container's
		// entrypoint. It will be translated to Docker's --entrypoint
		// option while creating the container. The syntax is similar to
		// Dockerfile's ENTRYPOINT directive, where each shell token is a
		// separate string in the array.
		entrypoint?: [_, ...]
		pull_policy?: "always" | "never" | "if-not-present" | list.UniqueItems() & [..."always" | "never" | "if-not-present"] & [_, ...] | *"always"
	} | [...string]

	#services: [...strings.MinRunes(1) | {
		// Full name of the image that should be used. It should contain
		// the Registry part if needed.
		name: strings.MinRunes(1)

		// Command or script that should be executed as the container's
		// entrypoint. It will be translated to Docker's --entrypoint
		// option while creating the container. The syntax is similar to
		// Dockerfile's ENTRYPOINT directive, where each shell token is a
		// separate string in the array.
		entrypoint?:  [_, ...] & [...string]
		pull_policy?: "always" | "never" | "if-not-present" | list.UniqueItems() & [..."always" | "never" | "if-not-present"] & [_, ...] | *"always"

		// Command or script that should be used as the container's
		// command. It will be translated to arguments passed to Docker
		// after the image's name. The syntax is similar to Dockerfile's
		// CMD directive, where each shell token is a separate string in
		// the array.
		command?: [_, ...] & [...string]

		// Additional alias that can be used to access the service from
		// the job's container. Read Accessing the services for more
		// information.
		alias?: strings.MinRunes(1)
	}]

	#id_tokens: {
		{[=~".*" & !~"^()$"]: aud: string | list.UniqueItems() & [...string] & [_, ...]}
		...
	}

	#secrets: [string]: {
		vault: string | {
			engine: {
				name: string
				path: string
				...
			}
			path:  string
			field: string
			...
		}
		...
	}

	#before_script: [...(string | [...string]) & (string | [...])]

	#after_script: [...(string | [...string]) & (string | [...])]

	#rules: [...({
		if?:            #if
		changes?:       #changes
		exists?:        #exists
		variables?:     #rulesVariables
		when?:          #when
		start_in?:      #start_in
		allow_failure?: #allow_failure
	} | strings.MinRunes(1) | [...string]) & (string | [...] | {
		...
	})]

	#workflowName: strings.MinRunes(1) & strings.MaxRunes(255)

	#globalVariables: {
		{[=~".*" & !~"^()$"]: number | string | {
			value?: string, description?: string, expand?: bool
		}}
		...
	}

	#jobVariables: {
		{[=~".*" & !~"^()$"]: number | string | {
			value?: string, expand?: bool
		}}
		...
	}

	#rulesVariables: {
		{[=~".*" & !~"^()$"]: number | string}
		...
	}

	#if: string

	#changes: ({
		// List of file paths.
		paths: [...string]

		// Ref for comparing changes.
		compare_to?: string
	} | [...string]) & ([...] | {
		...
	})

	#exists: [...string]

	#timeout: strings.MinRunes(1)

	#start_in: strings.MinRunes(1)

	#allow_failure: bool | *false | {
		exit_codes: int
	} | {
		exit_codes: list.UniqueItems() & [_, ...] & [...int]
	}

	#when: "on_success" | "on_failure" | "always" | "never" | "manual" | "delayed" | *"on_success"

	#cache: null | bool | number | string | [...] | {
		key?: string | {
			files?:  list.MaxItems(2) & [...string] & [_, ...]
			prefix?: string
			...
		}
		paths?: [...string]
		policy?:    "pull" | "push" | "pull-push" | *"pull-push"
		untracked?: bool | *false
		when?:      "on_success" | "on_failure" | "always" | *"on_success"
		...
	}

	#filter_refs: [...("branches" | "tags" | "api" | "external" | "pipelines" | "pushes" | "schedules" | "triggers" | "web" | string) & string]

	#filter: null | #filter_refs | {
		refs?: #filter_refs

		// Filter job based on if Kubernetes integration is active.
		kubernetes?: "active"
		variables?: [...string]

		// Filter job creation based on files that were modified in a git
		// push.
		changes?: [...string]
	}

	#retry: #retry_max | {
		max?:  #retry_max
		when?: #retry_errors | [...#retry_errors]
	}

	#retry_max: int & >=0 & <=2 | *0

	#retry_errors: "always" | "unknown_failure" | "script_failure" | "api_failure" | "stuck_or_timeout_failure" | "runner_system_failure" | "runner_unsupported" | "stale_schedule" | "job_execution_timeout" | "archived_failure" | "unmet_prerequisites" | "scheduler_failure" | "data_integrity_failure"

	#interruptible: bool | *false

	#job: #job_template

	#job_template: ({
		when:     "delayed"
		start_in: _
		...
	} | {
		when?: _
		...
	}) & {
		image?:         #image
		services?:      #services
		before_script?: #before_script
		after_script?:  #after_script
		rules?:         #rules
		variables?:     #jobVariables
		cache?:         #cache
		id_tokens?:     #id_tokens
		secrets?:       #secrets
		script?:        strings.MinRunes(1) | [...(string | [...string]) & (string | [...])] & [_, ...]

		// Define what stage the job will run in.
		stage?: (strings.MinRunes(1) | [...string]) & (string | [...])

		// Job will run *only* when these filtering options match.
		only?: #filter

		// The name of one or more jobs to inherit configuration from.
		extends?: string | [...string] & [_, ...]

		// The list of jobs in previous stages whose sole completion is
		// needed to start the current job.
		needs?: [...string | {
			job:        string
			artifacts?: bool
			optional?:  bool
		} | {
			pipeline:   string
			job:        string
			artifacts?: bool
		} | {
			job:        string
			project:    string
			ref:        string
			artifacts?: bool
		}]

		// Job will run *except* for when these filtering options match.
		except?:        #filter
		tags?:          #tags
		allow_failure?: #allow_failure
		timeout?:       #timeout
		when?:          #when
		start_in?:      #start_in

		// Specify a list of job names from earlier stages from which
		// artifacts should be loaded. By default, all previous artifacts
		// are passed. Use an empty array to skip downloading artifacts.
		dependencies?: [...string]
		artifacts?: #artifacts

		// Used to associate environment metadata with a deploy.
		// Environment can have a name and URL attached to it, and will
		// be displayed under /environments under the project.
		environment?: string | {
			// The name of the environment, e.g. 'qa', 'staging',
			// 'production'.
			name: strings.MinRunes(1)

			// When set, this will expose buttons in various places for the
			// current environment in Gitlab, that will take you to the
			// defined URL.
			url?: =~"^(https?://.+|\\$[A-Za-z]+)"

			// The name of a job to execute when the environment is about to
			// be stopped.
			on_stop?: string

			// Specifies what this job will do. 'start' (default) indicates
			// the job will start the deployment. 'prepare'/'verify'/'access'
			// indicates this will not affect the deployment. 'stop'
			// indicates this will stop the deployment.
			action?: "start" | "prepare" | "stop" | "verify" | "access" | *"start"

			// The amount of time it should take before Gitlab will
			// automatically stop the environment. Supports a wide variety of
			// formats, e.g. '1 week', '3 mins 4 sec', '2 hrs 20 min',
			// '2h20min', '6 mos 1 day', '47 yrs 6 mos and 4d', '3 weeks and
			// 2 days'.
			auto_stop_in?: string

			// Used to configure the kubernetes deployment for this
			// environment. This is currently not supported for kubernetes
			// clusters that are managed by Gitlab.
			kubernetes?: {
				// The kubernetes namespace where this environment should be
				// deployed to.
				namespace?: strings.MinRunes(1)
				...
			}

			// Explicitly specifies the tier of the deployment environment if
			// non-standard environment name is used.
			deployment_tier?: "production" | "staging" | "testing" | "development" | "other"
		}

		// Indicates that the job creates a Release.
		release?: {
			// The tag_name must be specified. It can refer to an existing Git
			// tag or can be specified by the user.
			tag_name: strings.MinRunes(1)

			// Message to use if creating a new annotated tag.
			tag_message?: string

			// Specifies the longer description of the Release.
			description: strings.MinRunes(1)

			// The Release name. If omitted, it is populated with the value of
			// release: tag_name.
			name?: string

			// If the release: tag_name doesn’t exist yet, the release is
			// created from ref. ref can be a commit SHA, another tag name,
			// or a branch name.
			ref?: string

			// The title of each milestone the release is associated with.
			milestones?: [...string]

			// The date and time when the release is ready. Defaults to the
			// current date and time if not defined. Should be enclosed in
			// quotes and expressed in ISO 8601 format.
			released_at?: =~"^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(?:Z|[+-][01]\\d:[0-5]\\d)$"
			assets?: {
				// Include asset links in the release.
				links: [...{
					// The name of the link.
					name: strings.MinRunes(1)

					// The URL to download a file.
					url: strings.MinRunes(1)

					// The redirect link to the url.
					filepath?: string

					// The content kind of what users can download via url.
					link_type?: "runbook" | "package" | "image" | "other"
				}] & [_, ...]
			}
		}

		// Must be a regular expression, optionally but recommended to be
		// quoted, and must be surrounded with '/'. Example: '/Code
		// coverage: \d+\.\d+/'
		coverage?: =~"^/.+/$"
		retry?:    #retry

		// Parallel will split up a single job into several, and provide
		// `CI_NODE_INDEX` and `CI_NODE_TOTAL` environment variables for
		// the running jobs.
		parallel?: int & >=2 & <=50 | *0 | {
			// Defines different variables for jobs that are running in
			// parallel.
			matrix: list.MaxItems(50) & [...{
				[string]: number | string | [...]
			}]
		}
		interruptible?: #interruptible

		// Limit job concurrency. Can be used to ensure that the Runner
		// will not run certain jobs simultaneously.
		resource_group?: string
		trigger?:        {
			// Path to the project, e.g. `group/project`, or
			// `group/sub-group/project`.
			project: =~"(?:\\S/\\S|\\$\\S+)"

			// The branch name that a downstream pipeline will use
			branch?: string

			// You can mirror the pipeline status from the triggered pipeline
			// to the source bridge job by using strategy: depend
			strategy?: "depend"

			// Specify what to forward to the downstream pipeline.
			forward?: {
				// Variables defined in the trigger job are passed to downstream
				// pipelines.
				yaml_variables?: bool | *true

				// Variables added for manual pipeline runs and scheduled
				// pipelines are passed to downstream pipelines.
				pipeline_variables?: bool | *false
			}
		} | {
			include?: =~"\\.ya?ml$" | list.MaxItems(3) & [...{
				// Relative path from local repository root (`/`) to the local
				// YAML file to define the pipeline configuration.
				local?: =~"\\.ya?ml$"
			} | {
				// Name of the template YAML file to use in the pipeline
				// configuration.
				template?: =~"\\.ya?ml$"
			} | {
				// Relative path to the generated YAML file which is extracted
				// from the artifacts and used as the configuration for
				// triggering the child pipeline.
				artifact: =~"\\.ya?ml$"

				// Job name which generates the artifact
				job: string
			} | {
				// Path to another private project under the same GitLab instance,
				// like `group/project` or `group/sub-group/project`.
				project: =~"(?:\\S/\\S|\\$\\S+)"

				// Branch/Tag/Commit hash for the target project.
				ref?: strings.MinRunes(1)

				// Relative path from repository root (`/`) to the pipeline
				// configuration YAML file.
				file: =~"\\.ya?ml$"
			}]

			// You can mirror the pipeline status from the triggered pipeline
			// to the source bridge job by using strategy: depend
			strategy?: "depend"

			// Specify what to forward to the downstream pipeline.
			forward?: {
				// Variables defined in the trigger job are passed to downstream
				// pipelines.
				yaml_variables?: bool | *true

				// Variables added for manual pipeline runs and scheduled
				// pipelines are passed to downstream pipelines.
				pipeline_variables?: bool | *false
			}
		} | =~"(?:\\S/\\S|\\$\\S+)"
		inherit?: {
			default?:   bool | [..."after_script" | "artifacts" | "before_script" | "cache" | "image" | "interruptible" | "retry" | "services" | "tags" | "timeout"]
			variables?: bool | [...string]
		}
	}

	#tags: [...(strings.MinRunes(1) | [...string]) & (string | [...])]
}
