package compose

import (
	"github.com/djmaze/shepherd/swarm"
)

const ComposeFileVersion string = "3.2"

type ComposeStack struct {
	Version  string
	Name     string                    "-"
	Services ComposeServiceCollection  ",inline"
	Volumes  []ComposeVolume           ",omitempty"
	Networks map[string]ComposeNetwork ",omitempty"
}

type ComposeServiceCollection struct {
	Services map[string]ComposeService
}

func ServiceCollectionFromSwarmServices(swarmServices []swarm.Service) ComposeServiceCollection {
	services := make(map[string]ComposeService)

	for _, swarmService := range swarmServices {
		composeService := NewService(swarmService)
		services[composeService.Name] = composeService
	}

	return ComposeServiceCollection{Services: services}
}

func (composeServiceCollection ComposeServiceCollection) Stacks() []ComposeStack {
	var stacks []ComposeStack

	for _, stackName := range composeServiceCollection.stackNames() {
		stack := composeServiceCollection.withStack(stackName).getStack(stackName)
		stacks = append(stacks, stack)
	}

	return stacks
}

func (composeServiceCollection ComposeServiceCollection) stackNames() []string {
	var result []string

	for _, composeService := range composeServiceCollection.Services {
		result = append(result, composeService.StackName)
	}

	return result
}

func (composeServiceCollection ComposeServiceCollection) withStack(stackName string) ComposeServiceCollection {
	servicesByName := make(map[string]ComposeService)

	for _, composeService := range composeServiceCollection.Services {
		if composeService.StackName == stackName {
			servicesByName[composeService.Name] = composeService
		}
	}

	return ComposeServiceCollection{Services: servicesByName}
}

func (composeServiceCollection ComposeServiceCollection) getStack(stackName string) ComposeStack {
	return ComposeStack{
		Version:  ComposeFileVersion,
		Name:     stackName,
		Services: composeServiceCollection,
		Networks: composeServiceCollection.getNetworks(),
	}
}

func (composeServiceCollection ComposeServiceCollection) getNetworks() map[string]ComposeNetwork {
	networks := make(map[string]ComposeNetwork)

	for _, composeService := range composeServiceCollection.Services {
		for _, composeNetwork := range composeService.ComposeNetworks() {
			networks[composeNetwork.Name] = composeNetwork
		}
	}

	return networks
}
