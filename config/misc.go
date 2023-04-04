package config

import (
	"os"
	"regexp"
	"strings"
)

// GetWorkspaceToken returns the workspace token provided in the environment variables
// Env variable CONFIG_BACKEND_TOKEN is deprecating soon
// WORKSPACE_TOKEN is newly introduced. This will override CONFIG_BACKEND_TOKEN
func GetWorkspaceToken() string {
	token := GetString("WORKSPACE_TOKEN", "")
	if token != "" && token != "<your_token_here>" {
		return token
	}
	return GetString("CONFIG_BACKEND_TOKEN", "")
}

// GetNamespaceIdentifier returns value stored in KUBE_NAMESPACE env var or "none" if empty
func GetNamespaceIdentifier() string {
	k8sNamespace := GetKubeNamespace()
	if k8sNamespace != "" {
		return k8sNamespace
	}
	return "none"
}

// GetKubeNamespace returns value stored in KUBE_NAMESPACE env var
func GetKubeNamespace() string {
	return os.Getenv("KUBE_NAMESPACE")
}

func GetInstanceID() string {
	instance := GetString("INSTANCE_ID", "")
	instanceArr := strings.Split(instance, "-")
	length := len(instanceArr)
	regexGwHa := regexp.MustCompile(`^.*-gw-ha-\d+-\w+-\w+$`)
	regexGwNonHaOrProcessor := regexp.MustCompile(`^.*-\d+$`)
	// This handles 2 kinds of server instances
	// a) Processor OR Gateway running in non HA mod where the instance name ends with the index
	// b) Gateway running in HA mode, where the instance name is of the form *-gw-ha-<index>-<statefulset-id>-<pod-id>
	if (regexGwHa.MatchString(instance)) && (length > 3) {
		return instanceArr[length-3]
	} else if (regexGwNonHaOrProcessor.MatchString(instance)) && (length > 1) {
		return instanceArr[length-1]
	}
	return ""
}

func GetReleaseName() string {
	return os.Getenv("RELEASE_NAME")
}
