package main

import (
	"errors"
	"strings"
	"github.com/aymerick/raymond"
)

func (renderContext *RenderContext) Build(deployerSpec DeployerSpec, objects []map[string]interface{}) error {
	envSlug := MakeUrlSlug(deployerSpec.Env, DNS_MAX_LENGTH)
	branchSlug := MakeUrlSlug(deployerSpec.Branch, DNS_MAX_LENGTH)

	renderContext.Namespace = deployerSpec.Namespace
	renderContext.Env = envSlug
	renderContext.Branch = branchSlug
	renderContext.Containers = make(map[string]RenderContextContainer)
	renderContext.Objects = make(map[string]map[string]RenderContextEnvAwareObject)

	for _, container := range deployerSpec.Containers {
		if container.Image == "" {
			continue
		}

		containerName := container.Image + ":" + deployerSpec.TagVersion
		renderContext.Containers[container.Id] = RenderContextContainer{
			Name:  containerName,
			Image: containerName,
		}
	}

	for _, object := range objects {
		kind := object["kind"].(string)
		var metadata = object["metadata"].(map[interface{}]interface{})
		name := metadata["name"].(string)
		envAwareName, err := buildEnvAwareObjectName(kind, name, renderContext)

		if err != nil {
			return err
		}

		if renderContext.Objects[kind] == nil {
			renderContext.Objects[kind] = make(map[string]RenderContextEnvAwareObject)
		}

		renderContext.Objects[kind][name] = RenderContextEnvAwareObject{
			Name: envAwareName,
		}
	}

	renderContext.DeployerSpec = deployerSpec

	return nil
}

func (renderContext *RenderContext) Render(templates []string) (string, error) {
	renderedTemplates := make([]string, 0)

	for _, template := range templates {

		ctx := map[string]*RenderContext{
			"context": renderContext,
		}

		renderedTemplate, err := raymond.Render(template, ctx)

		if err != nil {
			return "", err
		}

		renderedTemplates = append(renderedTemplates, renderedTemplate)
	}

	return strings.Join(renderedTemplates, "\n---\n"), nil
}

func buildEnvAwareObjectName(kind string, objectName string, renderContext *RenderContext) (string, error) {
	sanitizedObjectName := MakeUrlSlug(objectName, DNS_MAX_LENGTH)

	if (len(sanitizedObjectName) + 1 + len(renderContext.Env)) > DNS_MAX_LENGTH {
		return "", errors.New("object name is too long: " + renderContext.Env + "-" + sanitizedObjectName)
	}

	envAwareObjectName := renderContext.Env + "-" + sanitizedObjectName

	return envAwareObjectName, nil
}

type RenderContext struct {
	Containers   map[string]RenderContextContainer
	Objects      map[string]map[string]RenderContextEnvAwareObject
	Env          string
	Branch       string
	Namespace    string
	DeployerSpec DeployerSpec
}

type RenderContextContainer struct {
	Name  string
	Image string // Only for bc reasons
}

type RenderContextEnvAwareObject struct {
	Name string
}
