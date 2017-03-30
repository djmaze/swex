package swarm

import (
	"regexp"

	dockertypes "github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Client interface {
	ListServices() ([]Service, error)
	PullImage(name string) (Image, error)
	UpdateServiceImage(service Service, name string) error
}

type Service struct {
	ID    string
	Image Image
}

type Image struct {
	Name string
	Tag  string
	Hash string
}

type swarmClient struct {
	api *dockerclient.Client
}

func NewClient() (Client, error) {
	cli, err := dockerclient.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return swarmClient{api: cli}, nil
}

func (client swarmClient) ListServices() ([]Service, error) {
	bg := context.Background()

	services, err := client.api.ServiceList(bg, dockertypes.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	var swarmServices []Service

	for _, dockerService := range services {
		parts := regexp.MustCompile("[:@]").Split(dockerService.Spec.TaskTemplate.ContainerSpec.Image, 3)
		image := Image{Name: parts[0], Tag: parts[1], Hash: parts[2]}
		swarmService := Service{ID: dockerService.ID, Image: image}
		swarmServices = append(swarmServices, swarmService)
	}

	return swarmServices, err
}

func (client swarmClient) PullImage(name string) (Image, error) {
	return Image{}, nil
}

func (client swarmClient) UpdateServiceImage(service Service, name string) error {
	return nil
}
