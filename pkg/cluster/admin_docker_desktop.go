package cluster

import (
	"context"
	"fmt"

	"github.com/tilt-dev/localregistry-go"

	"github.com/tilt-dev/ctlptl/pkg/api"
	"github.com/tilt-dev/ctlptl/pkg/docker"
)

// The DockerDesktop manages the Kubernetes cluster for DockerDesktop.
// This is a bit different than the other admins, due to the overlap
//
type dockerDesktopAdmin struct {
	os   string
	host string
}

func newDockerDesktopAdmin(host string, os string) *dockerDesktopAdmin {
	return &dockerDesktopAdmin{os: os, host: host}
}

func (a *dockerDesktopAdmin) EnsureInstalled(ctx context.Context) error { return nil }
func (a *dockerDesktopAdmin) Create(ctx context.Context, desired *api.Cluster, registry *api.Registry) error {
	if registry != nil {
		return fmt.Errorf("ctlptl currently does not support connecting a registry to docker-desktop")
	}

	isLocalDockerDesktop := docker.IsLocalDockerDesktop(a.host, a.os)
	if !isLocalDockerDesktop {
		return fmt.Errorf("docker-desktop clusters are only available on a local Docker Desktop. Current DOCKER_HOST: %s",
			a.host)
	}

	return nil
}

func (a *dockerDesktopAdmin) LocalRegistryHosting(ctx context.Context, desired *api.Cluster, registry *api.Registry) (*localregistry.LocalRegistryHostingV1, error) {
	return nil, nil
}

func (a *dockerDesktopAdmin) Delete(ctx context.Context, config *api.Cluster) error {
	isLocalDockerHost := docker.IsLocalDockerDesktop(a.host, a.os)
	if !isLocalDockerHost {
		return fmt.Errorf("docker-desktop cannot be deleted from DOCKER_HOST: %s", a.host)
	}

	client, err := NewDockerDesktopClient()
	if err != nil {
		return err
	}

	err = client.ResetCluster(ctx)
	if err != nil {
		return err
	}

	settings, err := client.settings(ctx)
	if err != nil {
		return err
	}

	changed, err := client.setK8sEnabled(settings, false)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	return client.writeSettings(ctx, settings)
}
