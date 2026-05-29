package security

import (
	"strings"
	"testing"
)

func TestValidatorAccessLevels(t *testing.T) {
	tests := []struct {
		name        string
		accessLevel AccessLevel
		command     string
		shouldErr   bool
		errContains string
	}{
		// ReadOnly access level tests
		{"ReadOnly - get pods", AccessLevelReadOnly, "kubectl get pods", false, ""},
		{"ReadOnly - describe pod", AccessLevelReadOnly, "kubectl describe pod mypod", false, ""},
		{"ReadOnly - delete pod", AccessLevelReadOnly, "kubectl delete pod mypod", true, "read-only mode"},
		{"ReadOnly - create deployment", AccessLevelReadOnly, "kubectl create deployment nginx --image=nginx", true, "read-only mode"},
		{"ReadOnly - cordon node", AccessLevelReadOnly, "kubectl cordon node1", true, "read-only mode"},

		// ReadWrite access level tests
		{"ReadWrite - get pods", AccessLevelReadWrite, "kubectl get pods", false, ""},
		{"ReadWrite - delete pod", AccessLevelReadWrite, "kubectl delete pod mypod", false, ""},
		{"ReadWrite - create deployment", AccessLevelReadWrite, "kubectl create deployment nginx --image=nginx", false, ""},
		{"ReadWrite - cordon node", AccessLevelReadWrite, "kubectl cordon node1", true, "admin operations"},
		{"ReadWrite - drain node", AccessLevelReadWrite, "kubectl drain node1", true, "admin operations"},

		// Admin access level tests
		{"Admin - get pods", AccessLevelAdmin, "kubectl get pods", false, ""},
		{"Admin - delete pod", AccessLevelAdmin, "kubectl delete pod mypod", false, ""},
		{"Admin - create deployment", AccessLevelAdmin, "kubectl create deployment nginx --image=nginx", false, ""},
		{"Admin - cordon node", AccessLevelAdmin, "kubectl cordon node1", false, ""},
		{"Admin - drain node", AccessLevelAdmin, "kubectl drain node1", false, ""},

		// proxy access level tests
		{"ReadOnly - proxy blocked", AccessLevelReadOnly, "kubectl proxy --port=8001", true, "read-only mode"},
		{"ReadWrite - proxy allowed", AccessLevelReadWrite, "kubectl proxy --port=8001", false, ""},
		{"Admin - proxy allowed", AccessLevelAdmin, "kubectl proxy --port=8001", false, ""},

		// Config operations tests
		{"ReadOnly - config current-context", AccessLevelReadOnly, "config current-context", false, ""},
		{"ReadOnly - config get-contexts", AccessLevelReadOnly, "config get-contexts", false, ""},
		{"ReadOnly - config use-context", AccessLevelReadOnly, "config use-context mycontext", true, "config write operations in read-only mode"},
		{"ReadOnly - config delete-cluster", AccessLevelReadOnly, "config delete-cluster mycluster", true, "config write operations in read-only mode"},
		{"ReadOnly - config delete-context", AccessLevelReadOnly, "config delete-context myctx", true, "config write operations in read-only mode"},
		{"ReadOnly - config delete-user", AccessLevelReadOnly, "config delete-user myuser", true, "config write operations in read-only mode"},
		{"ReadOnly - config rename-context", AccessLevelReadOnly, "config rename-context old new", true, "config write operations in read-only mode"},
		{"ReadOnly - config unset", AccessLevelReadOnly, "config unset users.foo", true, "config write operations in read-only mode"},
		{"ReadOnly - config set", AccessLevelReadOnly, "config set current-context myctx", true, "config write operations in read-only mode"},
		{"ReadOnly - config set-context", AccessLevelReadOnly, "config set-context myctx", true, "config write operations in read-only mode"},
		{"ReadOnly - config set-cluster", AccessLevelReadOnly, "config set-cluster mycluster", true, "config write operations in read-only mode"},
		{"ReadOnly - config set-credentials", AccessLevelReadOnly, "config set-credentials myuser", true, "config write operations in read-only mode"},

		{"ReadWrite - config current-context", AccessLevelReadWrite, "config current-context", false, ""},
		{"ReadWrite - config get-contexts", AccessLevelReadWrite, "config get-contexts", false, ""},
		{"ReadWrite - config use-context", AccessLevelReadWrite, "config use-context mycontext", false, ""},
		{"ReadWrite - config delete-cluster", AccessLevelReadWrite, "config delete-cluster mycluster", false, ""},
		{"ReadWrite - config delete-context", AccessLevelReadWrite, "config delete-context myctx", false, ""},
		{"ReadWrite - config delete-user", AccessLevelReadWrite, "config delete-user myuser", false, ""},
		{"ReadWrite - config rename-context", AccessLevelReadWrite, "config rename-context old new", false, ""},
		{"ReadWrite - config unset", AccessLevelReadWrite, "config unset users.foo", false, ""},
		{"ReadWrite - config set", AccessLevelReadWrite, "config set current-context myctx", false, ""},
		{"ReadWrite - config set-context", AccessLevelReadWrite, "config set-context myctx", false, ""},
		{"ReadWrite - config set-cluster", AccessLevelReadWrite, "config set-cluster mycluster", false, ""},
		{"ReadWrite - config set-credentials", AccessLevelReadWrite, "config set-credentials myuser", false, ""},

		{"Admin - config current-context", AccessLevelAdmin, "config current-context", false, ""},
		{"Admin - config get-contexts", AccessLevelAdmin, "config get-contexts", false, ""},
		{"Admin - config use-context", AccessLevelAdmin, "config use-context mycontext", false, ""},
		{"Admin - config delete-cluster", AccessLevelAdmin, "config delete-cluster mycluster", false, ""},
		{"Admin - config delete-context", AccessLevelAdmin, "config delete-context myctx", false, ""},
		{"Admin - config delete-user", AccessLevelAdmin, "config delete-user myuser", false, ""},
		{"Admin - config rename-context", AccessLevelAdmin, "config rename-context old new", false, ""},
		{"Admin - config unset", AccessLevelAdmin, "config unset users.foo", false, ""},
		{"Admin - config set", AccessLevelAdmin, "config set current-context myctx", false, ""},
		{"Admin - config set-context", AccessLevelAdmin, "config set-context myctx", false, ""},
		{"Admin - config set-cluster", AccessLevelAdmin, "config set-cluster mycluster", false, ""},
		{"Admin - config set-credentials", AccessLevelAdmin, "config set-credentials myuser", false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			secConfig := NewSecurityConfig()
			secConfig.AccessLevel = tc.accessLevel
			validator := NewValidator(secConfig)

			err := validator.ValidateCommand(tc.command, CommandTypeKubectl)

			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for command %q with access level %s", tc.command, tc.accessLevel)
			} else if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error for command %q with access level %s: %v", tc.command, tc.accessLevel, err)
			} else if err != nil && tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("Error message should contain %q, got: %v", tc.errContains, err)
			}
		})
	}
}

func TestValidatorNamespaceRestriction(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("allowed-ns,another-ns")
	validator := NewValidator(secConfig)

	// Test allowed namespace
	err := validator.ValidateCommand("kubectl get pods -n allowed-ns", CommandTypeKubectl)
	if err != nil {
		t.Errorf("Allowed namespace should be accessible: %v", err)
	}

	// Test disallowed namespace
	err = validator.ValidateCommand("kubectl get pods -n disallowed-ns", CommandTypeKubectl)
	if err == nil {
		t.Error("Disallowed namespace should not be accessible")
	}

	// Test all namespaces restriction
	err = validator.ValidateCommand("kubectl get pods --all-namespaces", CommandTypeKubectl)
	if err == nil {
		t.Error("All namespaces should not be accessible when restrictions are in place")
	}
}

func TestNamespaceHandling(t *testing.T) {
	// Test namespace handling via public ValidateCommand method

	// Setup validator with namespace restrictions
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("test-ns,another-ns,default")
	validator := NewValidator(secConfig)

	// Test cases for namespace handling
	tests := []struct {
		command   string
		shouldErr bool
		errMsg    string
	}{
		{"kubectl get pods -n test-ns", false, ""},
		{"kubectl get pods --namespace=another-ns", false, ""},
		// resource/name without an explicit -n no longer silently defaults
		// to "default" (that was a bypass when "default" was in the
		// allowlist). It must be rejected, the same as bare "kubectl get
		// pods", because pod is a namespaced resource and the actual
		// namespace would come from kubeconfig.
		{"kubectl get pod/mypod", true, "explicit -n"},
		{"kubectl get pods -n disallowed-ns", true, "denied by security configuration"},
		{"kubectl get pods --all-namespaces", true, "restricted by security configuration"},
		{"kubectl get pods -A", true, "restricted by security configuration"},
	}

	for _, tc := range tests {
		err := validator.ValidateCommand(tc.command, CommandTypeKubectl)

		if tc.shouldErr && err == nil {
			t.Errorf("ValidateCommand(%q) should have failed", tc.command)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("ValidateCommand(%q) should have succeeded, got: %v", tc.command, err)
		} else if err != nil && tc.shouldErr && !strings.Contains(err.Error(), tc.errMsg) {
			t.Errorf("ValidateCommand(%q) error message mismatch, got: %v, want: %v", tc.command, err, tc.errMsg)
		}
	}
}

func TestReadOperationsValidation(t *testing.T) {
	// Test read operations validation through public API
	secConfig := NewSecurityConfig()
	secConfig.AccessLevel = AccessLevelReadOnly
	validator := NewValidator(secConfig)

	// Test cases for read operations
	tests := []struct {
		command     string
		commandType string
		shouldErr   bool
	}{
		{"kubectl get pods", CommandTypeKubectl, false},
		{"kubectl describe pod mypod", CommandTypeKubectl, false},
		{"kubectl delete pod mypod", CommandTypeKubectl, true},
		{"kubectl create namespace test", CommandTypeKubectl, true},
		{"helm list", CommandTypeHelm, false},
		{"helm status release", CommandTypeHelm, false},
		{"helm install chart", CommandTypeHelm, true},
		{"helm uninstall release", CommandTypeHelm, true},
		{"cilium status", CommandTypeCilium, false},
		{"cilium endpoint list", CommandTypeCilium, false}, // "endpoint" is in CiliumReadOperations
		{"cilium install", CommandTypeCilium, true},
		{"cilium hubble enable", CommandTypeCilium, false},
		{"hubble status", CommandTypeHubble, false},
		{"hubble observe", CommandTypeHubble, false},
		{"hubble list nodes", CommandTypeHubble, false},
		{"hubble config", CommandTypeHubble, false}, // "config" is in HubbleReadOperations
	}

	for _, tc := range tests {
		err := validator.ValidateCommand(tc.command, tc.commandType)

		if tc.shouldErr && err == nil {
			t.Errorf("ValidateCommand(%q, %q) should have failed", tc.command, tc.commandType)
		} else if !tc.shouldErr && err != nil {
			t.Errorf("ValidateCommand(%q, %q) should have succeeded, got: %v", tc.command, tc.commandType, err)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	// Comprehensive test with multiple security configurations
	testCases := []struct {
		name        string
		accessLevel AccessLevel
		namespaces  string
		command     string
		commandType string
		shouldErr   bool
	}{
		{"Read operation in readonly mode", AccessLevelReadOnly, "", "kubectl get pods", CommandTypeKubectl, false},
		{"Write operation in readonly mode", AccessLevelReadOnly, "", "kubectl delete pods", CommandTypeKubectl, true},

		{"Command in allowed namespace", AccessLevelReadWrite, "ns1,ns2", "kubectl get pods -n ns1", CommandTypeKubectl, false},
		{"Command in disallowed namespace", AccessLevelReadWrite, "ns1,ns2", "kubectl get pods -n ns3", CommandTypeKubectl, true},

		{"All namespaces restricted", AccessLevelReadWrite, "ns1,ns2", "kubectl get pods --all-namespaces", CommandTypeKubectl, true},

		// Combined restrictions
		{"Read op in allowed ns with readonly", AccessLevelReadOnly, "ns1", "kubectl get pods -n ns1", CommandTypeKubectl, false},
		{"Read op in disallowed ns with readonly", AccessLevelReadOnly, "ns1", "kubectl get pods -n ns2", CommandTypeKubectl, true},
		{"Write op in allowed ns with readonly", AccessLevelReadOnly, "ns1", "kubectl delete pods -n ns1", CommandTypeKubectl, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secConfig := NewSecurityConfig()
			secConfig.AccessLevel = tc.accessLevel
			if tc.namespaces != "" {
				secConfig.SetAllowedNamespaces(tc.namespaces)
			}

			validator := NewValidator(secConfig)
			err := validator.ValidateCommand(tc.command, tc.commandType)

			if tc.shouldErr && err == nil {
				t.Errorf("ValidateCommand should have failed")
			} else if !tc.shouldErr && err != nil {
				t.Errorf("ValidateCommand should have succeeded, got: %v", err)
			}
		})
	}
}

func TestValidateGlobalFlags(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.AccessLevel = AccessLevelReadOnly
	validator := NewValidator(secConfig)

	tests := []struct {
		name        string
		command     string
		commandType string
		shouldErr   bool
	}{
		// kubectl blocked flags
		{"kubectl --server= blocked", "kubectl get pods --server=https://attacker.com:8443", CommandTypeKubectl, true},
		{"kubectl --token= blocked", "kubectl get pods --token=abc123", CommandTypeKubectl, true},
		{"kubectl --kubeconfig= blocked", "kubectl get pods --kubeconfig=/tmp/evil", CommandTypeKubectl, true},
		{"kubectl --context= blocked", "kubectl get pods --context=evil-ctx", CommandTypeKubectl, true},
		{"kubectl --certificate-authority= blocked", "kubectl get pods --certificate-authority=/tmp/ca.crt", CommandTypeKubectl, true},
		{"kubectl --client-certificate= blocked", "kubectl get pods --client-certificate=/tmp/cert", CommandTypeKubectl, true},
		{"kubectl --client-key= blocked", "kubectl get pods --client-key=/tmp/key", CommandTypeKubectl, true},
		{"kubectl --insecure-skip-tls-verify blocked", "kubectl get pods --insecure-skip-tls-verify", CommandTypeKubectl, true},
		{"kubectl --as= blocked", "kubectl get pods --as=admin", CommandTypeKubectl, true},
		{"kubectl --as-group= blocked", "kubectl get pods --as-group=system:masters", CommandTypeKubectl, true},
		// kubectl normal flags allowed
		{"kubectl -n flag allowed", "kubectl get pods -n default", CommandTypeKubectl, false},
		{"kubectl -o flag allowed", "kubectl get pods -o yaml", CommandTypeKubectl, false},
		{"kubectl --namespace= allowed", "kubectl get pods --namespace=default", CommandTypeKubectl, false},
		// helm blocked flags
		{"helm --kube-apiserver= blocked", "helm list --kube-apiserver=https://attacker.com:8443", CommandTypeHelm, true},
		{"helm --kube-token= blocked", "helm list --kube-token=abc123", CommandTypeHelm, true},
		{"helm --kube-ca-file= blocked", "helm list --kube-ca-file=/tmp/ca.crt", CommandTypeHelm, true},
		{"helm --kube-context= blocked", "helm list --kube-context=evil", CommandTypeHelm, true},
		{"helm --kubeconfig= blocked", "helm list --kubeconfig=/tmp/evil", CommandTypeHelm, true},
		{"helm --kube-insecure-skip-tls-verify blocked", "helm list --kube-insecure-skip-tls-verify", CommandTypeHelm, true},
		// helm normal flags allowed
		{"helm -n flag allowed", "helm list -n default", CommandTypeHelm, false},
		// cilium/hubble not affected
		{"cilium with --server not blocked", "cilium status --server=evil", CommandTypeCilium, false},
		// Whitespace bypass regression — shlex splits on tab/CR/LF in addition
		// to space, so non-space separators must not slip past the substring scan.
		{"kubectl --server <tab> blocked", "kubectl get pods --server\thttps://attacker.example:8443", CommandTypeKubectl, true},
		{"kubectl --server <newline> blocked", "kubectl get pods --server\nhttps://attacker.example:8443", CommandTypeKubectl, true},
		{"kubectl --server <cr> blocked", "kubectl get pods --server\rhttps://attacker.example:8443", CommandTypeKubectl, true},
		{"kubectl --token <tab> blocked", "kubectl get pods --token\tabc123", CommandTypeKubectl, true},
		{"kubectl --token <newline> blocked", "kubectl get pods --token\nabc123", CommandTypeKubectl, true},
		{"kubectl --kubeconfig <tab> blocked", "kubectl get pods --kubeconfig\t/tmp/evil", CommandTypeKubectl, true},
		{"kubectl --as <tab> blocked", "kubectl get pods --as\tadmin", CommandTypeKubectl, true},
		{"helm --kube-apiserver <tab> blocked", "helm list --kube-apiserver\thttps://attacker.example:8443", CommandTypeHelm, true},
		{"helm --kubeconfig <newline> blocked", "helm list --kubeconfig\n/tmp/evil", CommandTypeHelm, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCommand(tc.command, tc.commandType)
			if tc.shouldErr && err == nil {
				t.Errorf("ValidateCommand(%q) should have been blocked", tc.command)
			} else if !tc.shouldErr && err != nil {
				t.Errorf("ValidateCommand(%q) should have been allowed, got: %v", tc.command, err)
			}
		})
	}
}

func TestValidateGlobalFlagsAllAccessLevels(t *testing.T) {
	// Blocked global flags must be rejected regardless of access level
	accessLevels := []AccessLevel{AccessLevelReadOnly, AccessLevelReadWrite, AccessLevelAdmin}
	for _, level := range accessLevels {
		secConfig := NewSecurityConfig()
		secConfig.AccessLevel = level
		validator := NewValidator(secConfig)

		err := validator.ValidateCommand("kubectl get pods --server=https://attacker.com:8443", CommandTypeKubectl)
		if err == nil {
			t.Errorf("--server= should be blocked at access level %s", level)
		}
	}
}

func TestNamespaceBypassPrevention(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("default")
	validator := NewValidator(secConfig)

	tests := []struct {
		name    string
		command string
	}{
		{"duplicate -n flags", "kubectl get secrets -n default -n kube-system -o yaml"},
		{"mixed namespace flags", "kubectl get secrets -n default --namespace=kube-system"},
		{"reverse order bypass", "kubectl get secrets -n kube-system -n default"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCommand(tc.command, CommandTypeKubectl)
			if err == nil {
				t.Errorf("Command with multiple namespace flags should be rejected: %q", tc.command)
			}
		})
	}
}

// TestNamespaceNoFlagBypass covers the MSRC report where commands without an
// explicit -n / --namespace flag silently bypassed --allow-namespaces and
// executed in the kubeconfig current namespace. The expected behavior is:
//   - When an allowlist is configured, namespaced operations without -n are
//     rejected.
//   - Cluster-scoped / non-resource operations (version, cluster-info, config,
//     api-resources, help, etc.) remain allowed.
//   - Compact short-flag form `-nVALUE` is parsed and validated.
func TestNamespaceNoFlagBypass(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("production")
	secConfig.AccessLevel = AccessLevelReadWrite
	validator := NewValidator(secConfig)

	tests := []struct {
		name        string
		command     string
		commandType string
		shouldBlock bool
	}{
		// Explicit allowed namespace: pass through
		{"explicit allowed ns", "kubectl get pods -n production", CommandTypeKubectl, false},
		{"explicit allowed --namespace=", "kubectl get pods --namespace=production", CommandTypeKubectl, false},

		// Explicit disallowed namespace: blocked
		{"explicit disallowed ns", "kubectl get pods -n default", CommandTypeKubectl, true},
		{"--namespace=disallowed", "kubectl get pods --namespace=default", CommandTypeKubectl, true},
		{"--all-namespaces", "kubectl get pods --all-namespaces", CommandTypeKubectl, true},
		{"-A short form", "kubectl get pods -A", CommandTypeKubectl, true},

		// No-flag bypass: now blocked
		{"no flag get secrets", "kubectl get secrets", CommandTypeKubectl, true},
		{"no flag get pods", "kubectl get pods", CommandTypeKubectl, true},
		{"no flag create secret", "kubectl create secret generic pwned --from-literal=x=y", CommandTypeKubectl, true},
		{"no flag apply -f", "kubectl apply -f deployment.yaml", CommandTypeKubectl, true},
		{"no flag get configmaps", "kubectl get configmaps", CommandTypeKubectl, true},
		{"no flag describe secret", "kubectl describe secret mysecret", CommandTypeKubectl, true},
		{"no flag logs", "kubectl logs mypod", CommandTypeKubectl, true},
		{"no flag delete secret", "kubectl delete secret mysecret", CommandTypeKubectl, true},
		{"no flag scale", "kubectl scale deployment myapp --replicas=3", CommandTypeKubectl, true},
		{"no flag helm list", "helm list", CommandTypeHelm, true},
		{"no flag helm status", "helm status myapp", CommandTypeHelm, true},

		// Compact -nVALUE form: parsed and enforced
		{"compact -nVALUE disallowed", "kubectl get secrets -ndefault", CommandTypeKubectl, true},
		{"compact -nVALUE allowed", "kubectl get secrets -nproduction", CommandTypeKubectl, false},

		// Cluster-scoped / non-resource operations: allowed without -n
		{"version no ns", "kubectl version", CommandTypeKubectl, false},
		{"cluster-info no ns", "kubectl cluster-info", CommandTypeKubectl, false},
		{"api-resources no ns", "kubectl api-resources", CommandTypeKubectl, false},
		{"api-versions no ns", "kubectl api-versions", CommandTypeKubectl, false},
		{"config get-contexts no ns", "kubectl config get-contexts", CommandTypeKubectl, false},
		{"explain pods no ns", "kubectl explain pods", CommandTypeKubectl, false},
		{"helm version no ns", "helm version", CommandTypeHelm, false},
		{"helm repo list no ns", "helm repo list", CommandTypeHelm, false},
		{"helm search no ns", "helm search repo nginx", CommandTypeHelm, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCommand(tc.command, tc.commandType)
			if tc.shouldBlock && err == nil {
				t.Errorf("expected block, got allow: %q", tc.command)
			}
			if !tc.shouldBlock && err != nil {
				t.Errorf("expected allow, got block: %q: %v", tc.command, err)
			}
		})
	}
}

// TestNamespaceNoRestrictionsAllowsEmptyNamespace makes sure that when no
// --allow-namespaces is configured we keep the historical permissive behavior:
// commands without -n are accepted (kubectl uses the kubeconfig current ns).
func TestNamespaceNoRestrictionsAllowsEmptyNamespace(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.AccessLevel = AccessLevelReadWrite
	validator := NewValidator(secConfig)

	commands := []struct {
		cmd  string
		kind string
	}{
		{"kubectl get pods", CommandTypeKubectl},
		{"kubectl get secrets", CommandTypeKubectl},
		{"kubectl apply -f deployment.yaml", CommandTypeKubectl},
		{"helm list", CommandTypeHelm},
	}
	for _, c := range commands {
		if err := validator.ValidateCommand(c.cmd, c.kind); err != nil {
			t.Errorf("unrestricted config should allow %q, got: %v", c.cmd, err)
		}
	}
}

// TestNamespaceResourceSlashBypass — review comment #1 on PR #141.
// extractNamespaceFromCommand previously returned "default" for any command
// containing a resource/name token, so with an allowlist that contained
// "default" (a common configuration), commands like `kubectl get
// secret/mysecret` silently bypassed the allowlist and ran in the
// kubeconfig current namespace. Additionally, `kubectl get deploy/myapp -A`
// hit the resource/name branch BEFORE the --all-namespaces check ran, so it
// was evaluated as a single namespace rather than as all-namespaces.
func TestNamespaceResourceSlashBypass(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("default,production")
	secConfig.AccessLevel = AccessLevelReadWrite
	validator := NewValidator(secConfig)

	tests := []struct {
		name        string
		command     string
		shouldBlock bool
	}{
		// resource/name without -n must NOT silently default to "default"
		// just because "default" is in the allowlist.
		{"resource/name no ns", "kubectl get secret/mysecret", true},
		{"resource/name delete", "kubectl delete secret/mysecret", true},
		// --all-namespaces must take precedence over resource/name parsing.
		{"resource/name -A", "kubectl get deploy/myapp -A", true},
		{"resource/name --all-namespaces", "kubectl get deploy/myapp --all-namespaces", true},
		// Explicit ns is still respected.
		{"resource/name -n allowed", "kubectl get secret/mysecret -n production", false},
		{"resource/name -n denied", "kubectl get secret/mysecret -n kube-system", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCommand(tc.command, CommandTypeKubectl)
			if tc.shouldBlock && err == nil {
				t.Errorf("expected block, got allow: %q", tc.command)
			}
			if !tc.shouldBlock && err != nil {
				t.Errorf("expected allow, got block: %q: %v", tc.command, err)
			}
		})
	}
}

// TestNamespaceFlagInsideExecArgs — review comment #2 on PR #141.
// The old regex matched -n / --namespace anywhere in the string, including
// inside the inner arguments of `kubectl exec ... -- ...`. That meant:
//
//   - benign inner `-n` (e.g. `grep -n`) made the validator parse a wrong
//     namespace from arbitrary text, blocking legitimate commands;
//   - crafted payloads like `... -- sh -c '... -n production ...'` made the
//     validator believe the namespace was allowed, while the real exec ran
//     in the kubeconfig current namespace.
//
// The fix tokenizes with shlex and ignores everything after a free-standing
// "--" separator.
func TestNamespaceFlagInsideExecArgs(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("production")
	secConfig.AccessLevel = AccessLevelReadWrite
	validator := NewValidator(secConfig)

	// Inner `-n` in grep args must NOT be parsed as the kube namespace.
	// Because the command itself doesn't specify a namespace and exec is
	// namespaced, this is rejected by the "needs explicit -n" rule -- the
	// important thing is that it is NOT rejected with a "namespace foo is
	// denied" message that would betray the buggy parse.
	err := validator.ValidateCommand("kubectl exec mypod -- grep -n foo file", CommandTypeKubectl)
	if err == nil {
		t.Errorf("exec without -n should be rejected when allowlist is configured")
	} else if strings.Contains(err.Error(), "namespace 'foo'") {
		t.Errorf("inner '-n foo' must not be parsed as kube namespace: %v", err)
	}

	// Crafted payload: attacker-controlled inner string contains
	// "-n production". The real exec still runs in the kubeconfig current
	// namespace, so the validator must NOT be fooled into approving it.
	err = validator.ValidateCommand(
		"kubectl exec mypod -- sh -c 'do_evil -n production'",
		CommandTypeKubectl,
	)
	if err == nil {
		t.Errorf("crafted inner -n production must not bypass --allow-namespaces")
	}

	// Adding an outer explicit -n makes it legitimate. Inner -n is still ignored.
	err = validator.ValidateCommand(
		"kubectl exec mypod -n production -- grep -n foo file",
		CommandTypeKubectl,
	)
	if err != nil {
		t.Errorf("legitimate exec with outer -n production should be allowed, got: %v", err)
	}
}

// TestClusterScopedResourcesAllowedWithoutNamespace — review comment #3 on
// PR #141. The verb-keyed exempt set (`version`, `cluster-info`, ...) does
// not cover routine cluster-scoped read commands like `kubectl get nodes`,
// because they use the same `get`/`top`/`auth` verbs as namespaced reads.
// The fix adds a cluster-scoped resource set so these commands stay
// allowed without -n.
func TestClusterScopedResourcesAllowedWithoutNamespace(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("production")
	secConfig.AccessLevel = AccessLevelReadWrite
	validator := NewValidator(secConfig)

	allowed := []string{
		"kubectl get nodes",
		"kubectl get no",
		"kubectl get pv",
		"kubectl get persistentvolumes",
		"kubectl get namespaces",
		"kubectl get ns",
		"kubectl get clusterroles",
		"kubectl get clusterrolebindings",
		"kubectl get storageclass",
		"kubectl get crd",
		"kubectl get csr",
		"kubectl get priorityclasses",
		"kubectl top nodes",
		"kubectl describe node node1",
		"kubectl get nodes,pv",         // multi-type, all cluster-scoped
		"kubectl get clusterrole/admin", // resource/name on cluster-scoped type
	}
	for _, c := range allowed {
		if err := validator.ValidateCommand(c, CommandTypeKubectl); err != nil {
			t.Errorf("cluster-scoped read should be allowed without -n: %q: %v", c, err)
		}
	}

	// Mixed cluster-scoped + namespaced resources are NOT exempt.
	blocked := []string{
		"kubectl get nodes,pods",          // mixed via comma
		"kubectl get node/n1 pod/p1",      // mixed via multiple resource/name positionals
		"kubectl get pods",                // namespaced
		"kubectl describe pod mypod",      // namespaced
		"kubectl get pod/mypod",           // namespaced via resource/name
		"kubectl get widget/foo",          // unknown type -> assume namespaced
		"kubectl auth can-i get pods",     // verb arg, not a resource
		"kubectl logs mypod",              // logs always targets a pod (namespaced)
		"kubectl label node node1 env=x",  // mutation verbs are not in the exempt verb set
		"kubectl delete node node1",       // mutation verbs are not in the exempt verb set
	}
	for _, c := range blocked {
		if err := validator.ValidateCommand(c, CommandTypeKubectl); err == nil {
			t.Errorf("expected block (namespaced or mixed), got allow: %q", c)
		}
	}
}

// TestAllNamespacesPlusExplicitNamespaceIsAmbiguous — when a command
// contains both --all-namespaces and -n X, the intent is ambiguous and
// should be rejected rather than silently honoring one form.
func TestAllNamespacesPlusExplicitNamespaceIsAmbiguous(t *testing.T) {
	secConfig := NewSecurityConfig()
	secConfig.SetAllowedNamespaces("production")
	validator := NewValidator(secConfig)

	cases := []string{
		"kubectl get pods -A -n production",
		"kubectl get pods --all-namespaces -n production",
		"kubectl get pods --all-namespaces --namespace=production",
	}
	for _, c := range cases {
		err := validator.ValidateCommand(c, CommandTypeKubectl)
		if err == nil {
			t.Errorf("ambiguous all-namespaces+explicit must be rejected: %q", c)
		}
	}
}
