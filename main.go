package main

import (
	"os"
	"os/exec"
	"path/filepath"

	demo "github.com/saschagrunert/demo"
)

func main() {
     d := demo.New()

     d.Add(runPolicyWithKwctl(), "run-policy-with-kwctl", "Runs a given policy with kwctl")
     d.Add(runPolicyOnKubernetes(), "run-policy-on-kubernetes", "Runs a given policy on top of Kubernets")

     d.Run()
}

func runPolicyWithKwctl() *demo.Run {
	r := demo.NewRun(
		"Running policies with kwctl",
	)

	r.Step(demo.S(
		"List policies",
	), demo.S(
		"kwctl policies",
	))
	
	r.Step(demo.S(
		"Download the ingress policy",
	), demo.S(
		"kwctl pull registry://ghcr.io/kubewarden/policies/ingress:v0.1.10",
	))

	r.Step(demo.S(
		"List policies",
	), demo.S(
		"kwctl policies",
	))
	
	r.Step(demo.S(
		"Inspect the policy",
	), demo.S(
		"kwctl inspect registry://ghcr.io/kubewarden/policies/ingress:v0.1.10",
	))

	r.Step(demo.S(
		"Request: TLS is not enabled in one host",
	), demo.S(
		"bat ingress-policy/tls-not-enabled.json",
	))

	r.Step(demo.S(
		"Run policy without any configuration",
		"This should be allowed",
	), demo.S(
		"kwctl run --request-path ingress-policy/tls-not-enabled.json --settings-json '{}' registry://ghcr.io/kubewarden/policies/ingress:v0.1.10",
	))

	r.Step(demo.S(
		"Run policy with TLS enforcement configuration disabled",
		"This should be allowed",
	), demo.S(
		`kwctl run --request-path ingress-policy/tls-not-enabled.json --settings-json '{"requireTLS": false}' registry://ghcr.io/kubewarden/policies/ingress:v0.1.10`,
	))

	r.Step(demo.S(
		"Run policy with TLS enforcement configuration enabled",
		"The request is missing TLS in at least one host: this should be disallowed",
	), demo.S(
		`kwctl run --request-path ingress-policy/tls-not-enabled.json --settings-json '{"requireTLS": true}' registry://ghcr.io/kubewarden/policies/ingress:v0.1.10`,
	))

	r.Step(demo.S(
		"Request: TLS is enabled in all hosts",
	), demo.S(
		"bat ingress-policy/tls-enabled.json",
	))

	r.Step(demo.S(
		"Run policy with TLS enforcement configuration enabled",
		"The request has TLS enabled in all hosts: this should be allowed",
	), demo.S(
		`kwctl run --request-path ingress-policy/tls-enabled.json --settings-json '{"requireTLS": true}' registry://ghcr.io/kubewarden/policies/ingress:v0.1.10`,
	))

	r.Setup(setupKubewarden)
	r.Cleanup(cleanupKubewarden)

	return r
}

func runPolicyOnKubernetes() *demo.Run {
	r := demo.NewRun(
		"Running policies on Kubernetes",
	)

	r.Step(demo.S(
		"List policies",
	), demo.S(
		"reg ls registry.ereslibre.net",
	))

	r.Step(demo.S(
		"Ingress: Let's Encrypt staging",
	),demo.S(
		"bat safe-annotations-policy/letsencrypt-staging-ingress.yaml",
	))

	r.Step(demo.S(
		"Create an Ingress resource using the letsencrypt-staging cert-manager issuer",
		"This is allowed: our policy is not deployed on the cluster yet",
	), demo.S(
		"kubectl apply -f safe-annotations-policy/letsencrypt-staging-ingress.yaml",
	))

	r.Step(demo.S(
		"Policy: safe annotations",
	),demo.S(
		"bat safe-annotations-policy/safe-annotations-letsencrypt-production-policy.yaml",
	))

	r.Step(demo.S(
		"Delete the ingress resource and deploy our policy",
	), demo.S(
		"kubectl delete ingress --all -n longhorn-test &&",
		"kubectl apply -f safe-annotations-policy/safe-annotations-letsencrypt-production-policy.yaml",
	))

	r.Step(demo.S(
		"Wait for our policy to be active",
	), demo.S(
		// We wait for PolicyServerWebhookConfigurationReconciled condition, but we should probably have a PolicyActive condition.
		// Despite we have a `policyActive` boolean field on the `status` structure, this is not a condition, and cannot be waited with `kubectl` as we can do here.
		"kubectl wait --for=condition=PolicyServerWebhookConfigurationReconciled clusteradmissionpolicy safe-annotations-lets-encrypt-production",
	))

	r.StepCanFail(demo.S(
		"Create an Ingress resource using the letsencrypt-staging cert-manager issuer",
		"This is rejected",
	), demo.S(
		"kubectl apply -f safe-annotations-policy/letsencrypt-staging-ingress.yaml",
	))

	r.Step(demo.S(
		"Ingress: Let's Encrypt production",
	),demo.S(
		"bat safe-annotations-policy/letsencrypt-production-ingress.yaml",
	))

	r.Step(demo.S(
		"Create an Ingress resource using the letsencrypt-production cert-manager issuer",
		"This is allowed",
	), demo.S(
		"kubectl apply -f safe-annotations-policy/letsencrypt-production-ingress.yaml",
	))

	r.Setup(func() error {
		setupKubewarden()
		setupKubernetes()
		return nil
	})
	r.Cleanup(func() error {
		cleanupKubewarden()
		cleanupKubernetes()
		return nil
	})
	
	return r
}

func setupKubewarden() error {
	cleanupKubewarden()
	return nil
}

func cleanupKubewarden() error {
	os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".cache", "kubewarden"))
	return nil
}

func setupKubernetes() error {
	exec.Command("kubectl", "create", "namespace", "longhorn-test").Run()
	return nil
}

func cleanupKubernetes() error {
	exec.Command("kubectl", "delete", "namespace", "longhorn-test").Run()
	exec.Command("kubectl", "delete", "clusteradmissionpolicy", "--all").Run()
	return nil
}
