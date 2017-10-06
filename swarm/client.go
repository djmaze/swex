package swarm

import (
	"regexp"

	dockertypes "github.com/docker/docker/api/types"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Client interface {
	ListServices() ([]Service, error)
	ListNetworks() (map[string]Network, error)
}

type Service struct {
	ID                   string
	Namespace            string
	Name                 string
	Image                Image
	Mode                 swarmtypes.ServiceMode
	Labels               map[string]string
	Command              []string
	Env                  []string
	Args                 []string
	Mounts               []Mount
	Ports                []Port
	Networks             []Network
	PlacementConstraints []string
}

type Image struct {
	Name string
	Tag  string
	Hash string
}

type Mount struct {
	Type     string
	Target   string
	Source   interface{}
	ReadOnly bool
	// TODO more
}

type Port struct {
	Protocol      string
	TargetPort    uint32
	PublishedPort uint32
}

type Network struct {
	Name   string
	ID     string
	Driver string
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
		//dockerService, _, err = client.api.ServiceInspectWithRaw(bg, dockerService.ID)

		parts := regexp.MustCompile("[:@]").Split(dockerService.Spec.TaskTemplate.ContainerSpec.Image, 3)
		var image Image
		switch len(parts) {
		case 1:
			image = Image{Name: parts[0], Tag: "latest"}
		case 2:
			image = Image{Name: parts[0], Tag: parts[1]}
		case 3:
			image = Image{Name: parts[0], Tag: parts[1], Hash: parts[2]}
		}

		var mounts []Mount
		for _, dockerMount := range dockerService.Spec.TaskTemplate.ContainerSpec.Mounts {
			mount := Mount{
				Target:   dockerMount.Target,
				Source:   dockerMount.Source,
				Type:     string(dockerMount.Type),
				ReadOnly: dockerMount.ReadOnly,
			}
			mounts = append(mounts, mount)
		}

		var ports []Port
		for _, dockerPort := range dockerService.Spec.EndpointSpec.Ports {
			port := Port{
				Protocol:      string(dockerPort.Protocol),
				TargetPort:    dockerPort.TargetPort,
				PublishedPort: dockerPort.PublishedPort,
			}
			ports = append(ports, port)
		}

		var networks []Network
		dockerNetworks, err := client.ListNetworks()
		if err != nil {
			return nil, err
		}
		for _, dockerNetwork := range dockerService.Spec.TaskTemplate.Networks {
			network := dockerNetworks[dockerNetwork.Target]
			networks = append(networks, network)
		}

		placement := dockerService.Spec.TaskTemplate.Placement
		var constraints []string
		if placement != nil {
			constraints = placement.Constraints
		} else {
			constraints = make([]string, 0)
		}

		var labels map[string]string
		if dockerService.Spec.Labels != nil {
			labels = dockerService.Spec.Labels
		} else {
			labels = make(map[string]string)
		}

		namespace := labels["com.docker.stack.namespace"]

		swarmService := Service{
			ID:                   dockerService.ID,
			Namespace:            namespace,
			Name:                 dockerService.Spec.Name,
			Image:                image,
			Mode:                 dockerService.Spec.Mode,
			Labels:               labels,
			Command:              dockerService.Spec.TaskTemplate.ContainerSpec.Command,
			Env:                  dockerService.Spec.TaskTemplate.ContainerSpec.Env,
			Args:                 dockerService.Spec.TaskTemplate.ContainerSpec.Args,
			Mounts:               mounts,
			Ports:                ports,
			Networks:             networks,
			PlacementConstraints: constraints,
		}
		swarmServices = append(swarmServices, swarmService)
	}

	return swarmServices, err
}

func (client swarmClient) ListNetworks() (map[string]Network, error) {
	bg := context.Background()

	networks, err := client.api.NetworkList(bg, dockertypes.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	networkMap := make(map[string]Network)
	for _, dockerNetwork := range networks {
		networkMap[dockerNetwork.ID] = Network{
			Name:   dockerNetwork.Name,
			ID:     dockerNetwork.ID,
			Driver: dockerNetwork.Driver,
		}
	}

	return networkMap, nil
}
