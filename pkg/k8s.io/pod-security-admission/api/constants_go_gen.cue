// Code generated by cue get go. DO NOT EDIT.

//cue:generate cue get go k8s.io/pod-security-admission/api

package api

#Level: string // #enumLevel

#enumLevel:
	#LevelPrivileged |
	#LevelBaseline |
	#LevelRestricted

#LevelPrivileged: #Level & "privileged"
#LevelBaseline:   #Level & "baseline"
#LevelRestricted: #Level & "restricted"

#VersionLatest: "latest"

#AuditAnnotationPrefix: "pod-security.kubernetes.io/"

_#labelPrefix:                 "pod-security.kubernetes.io/"
#EnforceLevelLabel:            "pod-security.kubernetes.io/enforce"
#EnforceVersionLabel:          "pod-security.kubernetes.io/enforce-version"
#AuditLevelLabel:              "pod-security.kubernetes.io/audit"
#AuditVersionLabel:            "pod-security.kubernetes.io/audit-version"
#WarnLevelLabel:               "pod-security.kubernetes.io/warn"
#WarnVersionLabel:             "pod-security.kubernetes.io/warn-version"
#ExemptionReasonAnnotationKey: "exempt"
#AuditViolationsAnnotationKey: "audit-violations"
#EnforcedPolicyAnnotationKey:  "enforce-policy"