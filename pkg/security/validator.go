package security

import (
	"strings"

	"github.com/google/shlex"
)

// Command type constants
const (
	CommandTypeKubectl = "kubectl"
	CommandTypeHelm    = "helm"
	CommandTypeCilium  = "cilium"
	CommandTypeHubble  = "hubble"
)

var (
	// KubectlBlockedGlobalFlags defines kubectl global flags that can redirect API traffic or inject credentials.
	// These are blocked at all access levels because they bypass the intent of access-level restrictions
	// by allowing traffic redirection and credential exfiltration.
	//
	// Entries are canonical flag names (without "=" or trailing space). Validation
	// tokenizes the command with shlex (matching the executor) and checks each
	// argv token against this set, so quoting / whitespace variants
	// (`--server"="X`, `--server\tX`, `--ser"ver"=X`, `--server\=X`) all reduce to
	// the same canonical token and are rejected.
	KubectlBlockedGlobalFlags = []string{
		"--server",
		"--token",
		"--kubeconfig",
		"--context",
		"--certificate-authority",
		"--client-certificate",
		"--client-key",
		"--insecure-skip-tls-verify",
		"--as",
		"--as-group",
		"--as-uid",
	}

	// HelmBlockedGlobalFlags defines helm global flags that can redirect API traffic or inject credentials.
	HelmBlockedGlobalFlags = []string{
		"--kube-apiserver",
		"--kube-token",
		"--kube-ca-file",
		"--kube-context",
		"--kubeconfig",
		"--kube-insecure-skip-tls-verify",
	}

	// KubectlReadOperations defines kubectl operations that don't modify state
	KubectlReadOperations = []string{
		"get", "describe", "explain", "logs", "top", "auth", "config",
		"cluster-info", "api-resources", "api-versions", "version", "diff",
		"completion", "help", "kustomize", "options", "plugin", "wait", "events",
	}

	// KubectlReadWriteOperations defines kubectl operations that modify state but are not admin operations
	KubectlReadWriteOperations = []string{
		"create", "delete", "apply", "expose", "run", "set", "rollout", "scale",
		"autoscale", "label", "annotate", "patch", "replace", "cp", "exec", "proxy",
	}

	// KubectlAdminOperations defines kubectl operations that require admin privileges
	KubectlAdminOperations = []string{
		"cordon", "uncordon", "drain", "taint", "certificate",
	}

	// HelmReadOperations defines helm operations that don't modify state
	HelmReadOperations = []string{
		"get", "history", "list", "show", "status", "search", "repo",
		"env", "version", "verify", "completion", "help",
	}

	// CiliumReadOperations defines cilium operations that don't modify state
	CiliumReadOperations = []string{
		"status", "version", "config", "help", "context", "connectivity",
		"endpoint", "identity", "ip", "map", "metrics", "monitor", "policy",
		"hubble", "bpf", "list", "observe", "service",
	}

	// HubbleReadOperations defines hubble operations that don't modify state
	HubbleReadOperations = []string{
		"status", "version", "help", "observe", "status", "list", "config",
	}

	// kubectlNamespaceExemptOperations are kubectl operations whose verbs are
	// inherently not bound to a single namespace and may therefore be executed
	// without an explicit -n flag even when --allow-namespaces is configured.
	// These cover cluster-info / version / kubeconfig-style introspection and
	// help commands. Verbs that can be either namespaced or cluster-scoped
	// (get, describe, top, auth, label, annotate, delete, patch, ...) are
	// NOT here -- those use kubectlClusterScopedResources instead.
	kubectlNamespaceExemptOperations = map[string]bool{
		"version":       true,
		"cluster-info":  true,
		"api-resources": true,
		"api-versions":  true,
		"config":        true,
		"completion":    true,
		"help":          true,
		"options":       true,
		"plugin":        true,
		"explain":       true,
	}

	// kubectlClusterScopedResources is the set of well-known cluster-scoped
	// resource types. When a kubectl command targets only resources in this
	// set, it does not act on any namespace and is exempt from
	// --allow-namespaces enforcement even without an explicit -n flag.
	//
	// Keys are lowercased, singular/plural variants and common short names are
	// all included so the lookup tolerates the forms users actually type.
	// Anything not in this set is treated as namespaced for safety.
	kubectlClusterScopedResources = map[string]bool{
		// Cluster-scoped core API
		"node": true, "nodes": true, "no": true,
		"namespace": true, "namespaces": true, "ns": true,
		"persistentvolume": true, "persistentvolumes": true, "pv": true,
		"componentstatus": true, "componentstatuses": true, "cs": true,

		// RBAC cluster-scoped
		"clusterrole": true, "clusterroles": true,
		"clusterrolebinding": true, "clusterrolebindings": true,

		// Storage cluster-scoped
		"storageclass": true, "storageclasses": true, "sc": true,
		"volumeattachment": true, "volumeattachments": true,
		"csinode": true, "csinodes": true,
		"csidriver": true, "csidrivers": true,

		// API & admission cluster-scoped
		"customresourcedefinition": true, "customresourcedefinitions": true, "crd": true, "crds": true,
		"apiservice": true, "apiservices": true,
		"mutatingwebhookconfiguration": true, "mutatingwebhookconfigurations": true,
		"validatingwebhookconfiguration": true, "validatingwebhookconfigurations": true,
		"validatingadmissionpolicy": true, "validatingadmissionpolicies": true,
		"validatingadmissionpolicybinding": true, "validatingadmissionpolicybindings": true,

		// Cert / scheduling / networking cluster-scoped
		"certificatesigningrequest": true, "certificatesigningrequests": true, "csr": true, "csrs": true,
		"priorityclass": true, "priorityclasses": true, "pc": true,
		"runtimeclass": true, "runtimeclasses": true,
		"ingressclass": true, "ingressclasses": true,
		"flowschema": true, "flowschemas": true,
		"prioritylevelconfiguration": true, "prioritylevelconfigurations": true,
	}

	// helmNamespaceExemptOperations mirror kubectlNamespaceExemptOperations
	// for helm. helm's namespaced commands (list/status/get/install/...) all
	// honor -n, but the entries below operate on local config / repo state.
	helmNamespaceExemptOperations = map[string]bool{
		"version":    true,
		"env":        true,
		"repo":       true,
		"search":     true,
		"completion": true,
		"help":       true,
		"verify":     true,
		"show":       true,
	}
)

// Validator handles validation of commands against security configuration
type Validator struct {
	secConfig *SecurityConfig
}

// NewValidator creates a new Validator instance with the given security configuration
func NewValidator(secConfig *SecurityConfig) *Validator {
	return &Validator{
		secConfig: secConfig,
	}
}

// ValidationError represents a security validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// getReadOperationsList returns the appropriate list of read operations based on command type
func (v *Validator) getReadOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlReadOperations
	case CommandTypeHelm:
		return HelmReadOperations
	case CommandTypeCilium:
		return CiliumReadOperations
	case CommandTypeHubble:
		return HubbleReadOperations
	default:
		return []string{}
	}
}

// getReadWriteOperationsList returns the appropriate list of read-write operations based on command type
func (v *Validator) getReadWriteOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlReadWriteOperations
	case CommandTypeHelm:
		// For now, assume helm write operations are same as read operations
		// This can be expanded when helm write operations are defined
		return []string{}
	case CommandTypeCilium:
		// For now, assume cilium write operations are same as read operations
		// This can be expanded when cilium write operations are defined
		return []string{}
	case CommandTypeHubble:
		// For now, assume hubble write operations are same as read operations
		// This can be expanded when hubble write operations are defined
		return []string{}
	default:
		return []string{}
	}
}

// getAdminOperationsList returns the appropriate list of admin operations based on command type
func (v *Validator) getAdminOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeKubectl:
		return KubectlAdminOperations
	case CommandTypeHelm:
		// For now, assume helm admin operations are not defined
		// This can be expanded when helm admin operations are defined
		return []string{}
	case CommandTypeCilium:
		// For now, assume cilium admin operations are not defined
		// This can be expanded when cilium admin operations are defined
		return []string{}
	case CommandTypeHubble:
		// For now, assume hubble admin operations are not defined
		// This can be expanded when hubble admin operations are defined
		return []string{}
	default:
		return []string{}
	}
}

// ValidateCommand validates a command against all security settings
func (v *Validator) ValidateCommand(command, commandType string) error {
	// Check for blocked global flags (credential/server redirection flags)
	if err := v.validateGlobalFlags(command, commandType); err != nil {
		return err
	}

	// Check access level restrictions
	if err := v.validateAccessLevel(command, commandType); err != nil {
		return err
	}

	// Check namespace scope restrictions
	if err := v.validateNamespaceScope(command, commandType); err != nil {
		return err
	}

	return nil
}

// validateGlobalFlags rejects commands that contain flags which can redirect API traffic
// or inject credentials, regardless of access level.
//
// The check operates on the same argv the executor will see: the command is
// tokenized with shlex (matching `pkg/command/command.go`), then each token
// is normalized to its canonical flag name (everything before the first `=`,
// lowercased) and compared to the blocked set. This closes the family of
// "tokenizer divergence" bypasses where the validator's raw-string scan and
// the executor's shlex tokenization disagree -- whitespace (tab/CR/LF),
// in-word quotes (`--server"="X`, `--ser"ver"=X`), and backslash escapes
// (`--server\=X`) all reduce to the same canonical flag token after shlex
// rejoins the pieces, so each variant is now caught.
func (v *Validator) validateGlobalFlags(command, commandType string) error {
	var blockedFlags []string
	switch commandType {
	case CommandTypeKubectl:
		blockedFlags = KubectlBlockedGlobalFlags
	case CommandTypeHelm:
		blockedFlags = HelmBlockedGlobalFlags
	default:
		return nil
	}

	blocked := make(map[string]struct{}, len(blockedFlags))
	for _, f := range blockedFlags {
		blocked[strings.ToLower(f)] = struct{}{}
	}

	tokens := tokenizeCommand(command)
	for _, t := range tokens {
		// Inspect only flag-shaped tokens. Positional args like resource
		// names cannot turn into a flag once shlex has split them out.
		if !strings.HasPrefix(t, "--") {
			continue
		}
		// Canonical name = everything before the first '=' (kubectl/helm
		// long flags use the `--flag=value` form). Bare booleans like
		// `--insecure-skip-tls-verify` have no '=', so the whole token is
		// the canonical name.
		name := t
		if i := strings.Index(t, "="); i >= 0 {
			name = t[:i]
		}
		name = strings.ToLower(name)
		if _, bad := blocked[name]; bad {
			return &ValidationError{Message: "Error: Global flag '" + name + "' is not allowed; it can redirect API traffic or inject credentials"}
		}
	}
	return nil
}

// validateAccessLevel validates if a command is allowed based on the configured access level
func (v *Validator) validateAccessLevel(command, commandType string) error {
	readOperations := v.getReadOperationsList(commandType)
	readWriteOperations := v.getReadWriteOperationsList(commandType)
	adminOperations := v.getAdminOperationsList(commandType)

	operation := v.extractOperationFromCommand(command, commandType)

	switch v.secConfig.AccessLevel {
	case AccessLevelReadOnly:
		// Special handling for config operations - check if it's a write operation
		if operation == "config" && v.isConfigWriteOperation(command) {
			return &ValidationError{Message: "Error: Cannot execute config write operations in read-only mode"}
		}
		if !v.isOperationInList(operation, readOperations) {
			return &ValidationError{Message: "Error: Cannot execute write or admin operations in read-only mode"}
		}
	case AccessLevelReadWrite:
		// Special handling for config operations - allow write config operations in readwrite mode
		if operation == "config" {
			return nil // All config operations are allowed in readwrite mode
		}
		if !v.isOperationInList(operation, readOperations) && !v.isOperationInList(operation, readWriteOperations) {
			// Check if it's an admin operation to provide better error message
			if v.isOperationInList(operation, adminOperations) {
				return &ValidationError{Message: "Error: Cannot execute admin operations in read-write mode"}
			}
			return &ValidationError{Message: "Error: Operation not allowed in read-write mode"}
		}
	case AccessLevelAdmin:
		// Admin level allows all operations (read, write, and admin), including all config operations
		if operation == "config" {
			return nil // All config operations are allowed in admin mode
		}
		if !v.isOperationInList(operation, readOperations) &&
			!v.isOperationInList(operation, readWriteOperations) &&
			!v.isOperationInList(operation, adminOperations) {
			return &ValidationError{Message: "Error: Unknown operation"}
		}
	default:
		return &ValidationError{Message: "Error: Invalid access level configuration"}
	}

	return nil
}

// Sentinel namespace tokens returned by namespace extraction.
const (
	namespaceTokenAmbiguous     = "__AMBIGUOUS_NAMESPACE__"
	namespaceTokenAllNamespaces = "*"
)

// tokenizeCommand splits a command string the same way the executor does
// (shlex), so the validator's view of `-n value`, quoting and the `--`
// separator matches what kubectl will actually receive. Falls back to
// strings.Fields when shlex rejects the input (unterminated quotes etc.)
// so validation still runs -- the executor will reject the broken input
// downstream.
func tokenizeCommand(command string) []string {
	if tokens, err := shlex.Split(command); err == nil {
		return tokens
	}
	return strings.Fields(command)
}

// splitArgsAtDoubleDash returns the slice of tokens up to (but excluding)
// a free-standing "--" separator. kubectl uses "--" to mark the start of
// arguments that should be passed to the child process (e.g. the inner
// `sh -c '...'` of `kubectl exec mypod -- ...`). Anything past "--" is not
// a kubectl flag and must not influence namespace / scope decisions.
func splitArgsAtDoubleDash(tokens []string) []string {
	for i, t := range tokens {
		if t == "--" {
			return tokens[:i]
		}
	}
	return tokens
}

// validateNamespaceScope validates if a command's namespace scope is allowed by security settings
func (v *Validator) validateNamespaceScope(command, commandType string) error {
	tokens := splitArgsAtDoubleDash(tokenizeCommand(command))

	namespace := extractNamespaceFromTokens(tokens)

	// Reject commands with multiple (ambiguous) namespace flags
	if namespace == namespaceTokenAmbiguous {
		return &ValidationError{Message: "Error: Command contains multiple namespace flags which is not allowed"}
	}

	hasRestrictions := len(v.secConfig.allowedNamespaces) > 0 || len(v.secConfig.allowedNamespacesRe) > 0

	// If command applies to all namespaces, and there are namespace restrictions
	if namespace == namespaceTokenAllNamespaces && hasRestrictions {
		return &ValidationError{Message: "Error: Access to all namespaces is restricted by security configuration"}
	}

	// If a namespace is specified, check if it's allowed
	if namespace != "" && namespace != namespaceTokenAllNamespaces {
		if !v.secConfig.IsNamespaceAllowed(namespace) {
			return &ValidationError{
				Message: "Error: Access to namespace '" + namespace + "' is denied by security configuration",
			}
		}
		return nil
	}

	// No explicit namespace was found. When allowlist restrictions are active,
	// commands without an explicit namespace would otherwise execute in the
	// kubeconfig current namespace, which silently bypasses the allowlist.
	// Reject these unless the command is inherently namespace-independent.
	if namespace == "" && hasRestrictions {
		if v.isCommandNamespaceExempt(tokens, commandType) {
			return nil
		}
		return &ValidationError{
			Message: "Error: Command does not specify a namespace; an explicit -n/--namespace flag is required when --allow-namespaces is configured",
		}
	}

	return nil
}

// isCommandNamespaceExempt reports whether a command may run without an
// explicit -n / --namespace flag even when --allow-namespaces is configured.
// A command is exempt when either:
//
//   - its operation verb is in the *NamespaceExemptOperations table (verbs
//     that never touch a namespace, e.g. kubectl version, helm repo, ...), or
//   - (kubectl only) every resource reference in the command targets a
//     cluster-scoped resource (kubectl get nodes, kubectl get pv, kubectl get
//     clusterrole/foo, ...).
//
// Anything ambiguous (unknown resource type, no resource type at all) is
// NOT exempt, so the default behavior is to require -n.
func (v *Validator) isCommandNamespaceExempt(tokens []string, commandType string) bool {
	operation := extractOperationFromTokens(tokens, commandType)

	switch commandType {
	case CommandTypeKubectl:
		if kubectlNamespaceExemptOperations[operation] {
			return true
		}
		return kubectlOnlyTargetsClusterScopedResources(tokens, operation)
	case CommandTypeHelm:
		return helmNamespaceExemptOperations[operation]
	default:
		// cilium / hubble are not gated by --allow-namespaces in this validator.
		return true
	}
}

// kubectlOnlyTargetsClusterScopedResources returns true iff the command
// references one or more resource types AND every referenced resource type
// is known cluster-scoped. Examples:
//
//	kubectl get nodes                       -> true  (nodes is cluster-scoped)
//	kubectl get nodes,pv                    -> true  (both cluster-scoped)
//	kubectl get clusterrole/admin           -> true  (resource/name form)
//	kubectl get pods                        -> false (pods is namespaced)
//	kubectl get nodes pods                  -> false (mixed)
//	kubectl get -f manifest.yaml            -> false (no resource type to inspect)
//	kubectl auth can-i get pods             -> false (verb arg, not a resource)
//	kubectl logs mypod                      -> false (mypod is a name, no type)
func kubectlOnlyTargetsClusterScopedResources(tokens []string, operation string) bool {
	// Only well-understood read/inspect verbs are eligible. Mutation verbs
	// like "label" or "delete" take resources too, but until we model the
	// argument grammar carefully we conservatively require -n for them.
	resourceVerbs := map[string]bool{
		"get": true, "describe": true, "top": true, "edit": true, "wait": true,
	}
	if !resourceVerbs[operation] {
		return false
	}

	resourceArgs := collectResourceArgs(tokens, operation)
	if len(resourceArgs) == 0 {
		return false
	}

	// kubectl get accepts two grammars:
	//   1. <type> <name1> <name2> ...        — type appears once, the rest are names
	//   2. <type>/<name> <type>/<name> ...   — every positional is its own type/name pair
	// If the first positional uses the resource/name form, we must inspect
	// EVERY positional (each carries its own type). Otherwise the first
	// positional carries the only type and the rest are names we can ignore.
	firstIsSlash := strings.Contains(resourceArgs[0], "/")

	argsToCheck := resourceArgs[:1]
	if firstIsSlash {
		argsToCheck = resourceArgs
	}

	for _, arg := range argsToCheck {
		// In slash-form pass, ignore positionals that are not themselves
		// slash form (they would be stray names left over from a malformed
		// command — be conservative and reject).
		if firstIsSlash && !strings.Contains(arg, "/") {
			return false
		}
		for _, rt := range splitResourceTypes(arg) {
			if !kubectlClusterScopedResources[strings.ToLower(rt)] {
				return false
			}
		}
	}
	return true
}

// collectResourceArgs returns the positional arguments that follow the
// operation verb, stopping at flags and at the first arg that is purely a
// name (no type). For "kubectl get nodes node1 -o yaml" it returns
// ["nodes", "node1"]; for "kubectl get nodes,pv" it returns ["nodes,pv"].
// Flag values (`-o yaml`, `-l app=x`, `--field-selector=...`) are skipped.
func collectResourceArgs(tokens []string, operation string) []string {
	flagsTakingValues := map[string]bool{
		"-o": true, "--output": true,
		"-l": true, "--selector": true,
		"--field-selector":      true,
		"--chunk-size":          true,
		"--show-managed-fields": true,
		"--sort-by":             true,
		"--template":            true,
		"--server-print":        true,
		"--for":                 true,
	}

	var out []string
	seenOp := false
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if !seenOp {
			if t == operation {
				seenOp = true
			}
			continue
		}
		if strings.HasPrefix(t, "-") {
			// Flag. Skip its value if it takes a separate argument and isn't using "=".
			if !strings.Contains(t, "=") && flagsTakingValues[t] && i+1 < len(tokens) {
				i++
			}
			continue
		}
		out = append(out, t)
	}
	return out
}

// splitResourceTypes turns a resource argument into the list of resource
// types it references. "nodes,pv" -> ["nodes", "pv"]; "clusterrole/admin"
// -> ["clusterrole"]; "node1" -> ["node1"] (just a name -- caller must
// decide what to do with bare names).
func splitResourceTypes(arg string) []string {
	parts := strings.Split(arg, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if i := strings.Index(p, "/"); i >= 0 {
			p = p[:i]
		}
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// isOperationInList checks if an operation is in the given list
func (v *Validator) isOperationInList(operation string, allowedOperations []string) bool {
	for _, allowed := range allowedOperations {
		if operation == allowed {
			return true
		}
	}
	return false
}

// extractOperationFromCommand extracts the operation from a command
func (v *Validator) extractOperationFromCommand(command, commandType string) string {
	return extractOperationFromTokens(splitArgsAtDoubleDash(tokenizeCommand(command)), commandType)
}

// extractOperationFromTokens returns the first positional token after the
// command name (skipping flags). For "kubectl get pods -n x" -> "get".
func extractOperationFromTokens(tokens []string, commandType string) string {
	for _, part := range tokens {
		if strings.HasPrefix(part, "-") {
			continue
		}
		if part == commandType {
			continue
		}
		return part
	}
	return ""
}

// isConfigWriteOperation checks if a config command is a write operation
func (v *Validator) isConfigWriteOperation(command string) bool {
	// Extract config subcommand
	cmdParts := strings.Fields(command)
	if len(cmdParts) < 2 || cmdParts[0] != "config" {
		return false
	}

	writeSubcommands := []string{
		"use-context",
		"set",
		"set-cluster",
		"set-context",
		"set-credentials",
		"unset",
		"delete-cluster",
		"delete-context",
		"delete-user",
		"rename-context",
	}

	subcommand := cmdParts[1]
	for _, writeOp := range writeSubcommands {
		if subcommand == writeOp {
			return true
		}
	}

	return false
}

// extractNamespaceFromTokens scans a pre-tokenized command for the
// namespace flag. Only flag tokens are inspected, so a `-n` that appears
// inside `kubectl exec ... -- grep -n foo file` (already trimmed by the
// caller via splitArgsAtDoubleDash) and a literal `-n` value within an
// `--option=value` style flag (token contains "=") are correctly ignored.
//
// Supported forms (all in a *single* flag token or a flag/value pair):
//
//	--namespace=value
//	--namespace value
//	-n=value
//	-n value
//	-nvalue          (compact pflag short form)
//
// All-namespaces forms: --all-namespaces / --all-namespaces=true|false / -A.
func extractNamespaceFromTokens(tokens []string) string {
	var (
		nsValues []string
		allNs    bool
	)

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		switch {
		case t == "--all-namespaces" || t == "-A":
			allNs = true
			continue
		case strings.HasPrefix(t, "--all-namespaces="):
			v := strings.TrimPrefix(t, "--all-namespaces=")
			if v != "false" && v != "0" {
				allNs = true
			}
			continue
		}

		// --namespace forms
		if t == "--namespace" {
			if i+1 < len(tokens) {
				nsValues = append(nsValues, tokens[i+1])
				i++
			}
			continue
		}
		if strings.HasPrefix(t, "--namespace=") {
			nsValues = append(nsValues, strings.TrimPrefix(t, "--namespace="))
			continue
		}

		// -n short forms
		if t == "-n" {
			if i+1 < len(tokens) {
				nsValues = append(nsValues, tokens[i+1])
				i++
			}
			continue
		}
		if strings.HasPrefix(t, "-n=") {
			nsValues = append(nsValues, strings.TrimPrefix(t, "-n="))
			continue
		}
		// Compact form `-nVALUE`. Restrict to tokens that start with literal
		// "-n" but are NOT another long flag like "--node-selector" and not
		// "-nA"/"-no" etc combined with other short flags. We require the
		// next character to be present and the token length > 2.
		if len(t) > 2 && strings.HasPrefix(t, "-n") && !strings.HasPrefix(t, "--") {
			nsValues = append(nsValues, t[2:])
			continue
		}
	}

	if len(nsValues) > 1 {
		return namespaceTokenAmbiguous
	}
	if len(nsValues) == 1 {
		if allNs {
			// Conflict: both -n X and --all-namespaces specified.
			return namespaceTokenAmbiguous
		}
		return nsValues[0]
	}
	if allNs {
		return namespaceTokenAllNamespaces
	}
	return ""
}
