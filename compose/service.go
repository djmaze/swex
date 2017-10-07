package compose

import (
	"strings"

	"github.com/djmaze/swex/swarm"
)

type ComposeService struct {
	SwarmService swarm.Service "-"
	Name         string        "-"
	StackName    string        "-"
	Image        string
	Command      []string      ",omitempty"
	Environment  []string      ",omitempty"
	Volumes      []swarm.Mount ",omitempty"
	Ports        []ExposedPort ",omitempty"
	Networks     []string      ",omitempty"
	Deploy       ComposeDeploy ",omitempty"
}

type ExposedPort struct {
	Target    uint32
	Published uint32
	Protocol  string ",omitempty"
	// TODO host mode
}

type ComposeDeploy struct {
	Mode      string
	Replicas  *uint64           ",omitempty"
	Labels    map[string]string ",omitempty"
	Placement ComposePlacement  ",omitempty"
}

type ComposePlacement struct {
	Constraints []string ",omitempty"
}

type ComposeVolume struct {
	Name string
}

type ComposeNetwork struct {
	Name     string "-"
	Driver   string
	External bool ",omitempty"
}

func NewService(swarmService swarm.Service) ComposeService {
	return ComposeService{
		SwarmService: swarmService,
		Name:         composeServiceName(swarmService),
		StackName:    composeStackName(swarmService),
		Image:        fullyQualifiedImage(swarmService),
		Command:      commandWithArgs(swarmService),
		Environment:  swarmService.Env,
		Volumes:      swarmService.Mounts,
		Ports:        exposedPorts(swarmService),
		Networks:     networkNames(swarmService),
		Deploy: ComposeDeploy{
			Mode:     modeString(swarmService),
			Replicas: replicaCount(swarmService),
			Labels:   composeLabels(swarmService),
			Placement: ComposePlacement{
				Constraints: swarmService.PlacementConstraints,
			},
		},
	}
}

func (composeService ComposeService) ComposeNetworks() map[string]ComposeNetwork {
	result := make(map[string]ComposeNetwork)

	for _, network := range composeService.SwarmService.Networks {
		external := !strings.Contains(network.Name, "_")
		result[network.Name] = ComposeNetwork{
			Name:     network.Name,
			Driver:   network.Driver,
			External: external,
		}
	}

	return result
}

func fullyQualifiedImage(swarmService swarm.Service) string {
	return swarmService.Image.Name + ":" + swarmService.Image.Tag
}

func exposedPorts(swarmService swarm.Service) []ExposedPort {
	var result []ExposedPort

	for _, servicePort := range swarmService.Ports {
		exposedPort := ExposedPort{
			Target:    servicePort.TargetPort,
			Published: servicePort.PublishedPort,
			Protocol:  servicePort.Protocol,
		}
		result = append(result, exposedPort)
	}

	return result
}

func commandWithArgs(swarmService swarm.Service) []string {
	var result []string

	result = append(result, swarmService.Command...)
	result = append(result, swarmService.Args...)

	return result
}

func networkNames(swarmService swarm.Service) []string {
	var result []string

	for _, network := range swarmService.Networks {
		result = append(result, network.Name)
	}

	return result
}

func modeString(swarmService swarm.Service) string {
	if swarmService.Mode.Global != nil {
		return "global"
	} else {
		return "replicated"
	}
}

func replicaCount(swarmService swarm.Service) *uint64 {
	if swarmService.Mode.Replicated != nil {
		return swarmService.Mode.Replicated.Replicas
	} else {
		return nil
	}
}

func composeServiceName(swarmService swarm.Service) string {
	if len(swarmService.Namespace) > 0 {
		return strings.TrimPrefix(swarmService.Name, swarmService.Namespace+"_")
	} else {
		return swarmService.Name
	}
}

func composeStackName(swarmService swarm.Service) string {
	if len(swarmService.Namespace) > 0 {
		return swarmService.Namespace
	} else {
		return swarmService.Name
	}
}

func composeLabels(swarmService swarm.Service) map[string]string {
	result := make(map[string]string)

	for key, value := range swarmService.Labels {
		if key != "com.docker.stack.namespace" {
			result[key] = value
		}
	}

	return result
}
