package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/MayaraCloud/terraform-provider-anthos/debug"
	"github.com/ghodss/yaml"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // This is needed for gcp auth
)

// Absolute Kubernetes API paths of the exclusivity artifacts
const (
	CRDAbspath string = "apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/memberships.hub.gke.io"
	CRAbspath         = "apis/hub.gke.io/v1/memberships/membership"
)

// GetMembershipCR get the Membership CR
func GetMembershipCR(ctx context.Context, auth Auth) (string, error) {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	object, err := kubeClient.RESTClient().Get().AbsPath(CRAbspath).DoRaw(ctx)
	if err != nil {
		// If there is no Membership CR we just return an empty string
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return "", nil
		}
		return "", fmt.Errorf("Getting the membership CR object: %w", err)
	}
	yamlObject, err := yaml.JSONToYAML(object)
	if err != nil {
		return "", fmt.Errorf("Transforming CR Manifest json to yaml: %w", err)
	}
	return string(string(yamlObject)), nil
}

// GetMembershipCRD get the Membership CRD
func GetMembershipCRD(ctx context.Context, auth Auth) (string, error) {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	object, err := kubeClient.RESTClient().Get().AbsPath(CRDAbspath).DoRaw(ctx)
	if err != nil {
		// If there is no Membership CRD we just return an empty string
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return "", nil
		}
		return "", fmt.Errorf("Getting the membership CRD object: %w", err)
	}
	yamlObject, err := yaml.JSONToYAML(object)
	if err != nil {
		return "", fmt.Errorf("Transforming CRD Manifest json to yaml: %w", err)
	}
	return string(string(yamlObject)), nil
}

// InstallExclusivityManifests applies the CRD and CR manifests in the cluster
// This will either install or upgrade them if already present
func InstallExclusivityManifests(ctx context.Context, auth Auth, CRDManifest string, CRManifest string) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	if CRDManifest != "" {
		debug.GoLog("InstallExclusivityManifests: installing CRD manifest")
		err = installRawArtifact(ctx, kubeClient, CRDAbspath, CRDManifest)
		if err != nil {
			return fmt.Errorf("Installing CRD: %w", err)
		}
	}
	if CRManifest != "" {
		debug.GoLog("InstallExclusivityManifests: installing CR manifest")
		err = installRawArtifact(ctx, kubeClient, CRAbspath, CRManifest)
		if err != nil {
			return fmt.Errorf("Installing CR: %w", err)
		}
	}

	debug.GoLog("InstallExclusivityManifests: received empty artifacts")
	return nil
}

func installRawArtifact(ctx context.Context, kubeClient *kubernetes.Clientset, absPath string, artifact string) error {
	JSONArtifact, err := yaml.YAMLToJSON([]byte(artifact))
	if err != nil {
		return fmt.Errorf("Converting yaml to json: %w", err)
	}
	_, err = kubeClient.RESTClient().Get().AbsPath(absPath).DoRaw(ctx)
	if err != nil {
		// If there is no artifact CREATE, otherwise, PATCH
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			// The CRD API requires a different absolute path on creation
			if absPath == CRDAbspath {
				absPath = "apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions"
			}
			debug.GoLog("installRawArtifact: installing the artifact " + absPath)

			// The creating api seems to only like JSON
			_, err = kubeClient.RESTClient().Post().Body(JSONArtifact).AbsPath(absPath).DoRaw(ctx)
			if err != nil {
				return fmt.Errorf("Error CREATING %v: %w", absPath, err)
			}
			return nil
		}
		// If the error is a not a "not found", then return it
		return fmt.Errorf("Getting %v: %w", absPath, err)
	}

	debug.GoLog("installRawArtifact: updating the artifact " + absPath)
	_, err = kubeClient.RESTClient().Patch(k8sTypes.ApplyPatchType).Body([]byte(artifact)).AbsPath(absPath).DoRaw(ctx)
	if err != nil {
		return fmt.Errorf("Error PATCHING %v: %w", absPath, err)
	}

	return nil
}

// DeleteArtifacts deletes the CRD and CR manifests in the cluster
func DeleteArtifacts(ctx context.Context, auth Auth) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	artifacts := []string{CRDAbspath, CRAbspath}
	for _, artifact := range artifacts {
		// Check if the artifact exists
		_, err := kubeClient.RESTClient().Get().AbsPath(artifact).DoRaw(ctx)
		if err != nil {
			// If there is no artifact no need to delete it
			if strings.Contains(err.Error(), "the server could not find the requested resource") {
				return nil
			}
			return fmt.Errorf("Getting the artifact before deleting: %w", err)
		}
		// Delete the resource
		_, err = kubeClient.RESTClient().Delete().AbsPath(artifact).DoRaw(ctx)
		if err != nil {
			return fmt.Errorf("Error DELETING %v: %w", artifact, err)
		}
	}

	return nil
}
