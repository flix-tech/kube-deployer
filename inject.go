package main

import (
	"crypto/md5"
	"encoding/hex"
)

const DEFAULT_REVISION_HISTORY_LIMIT = 3

type InjectContext struct {
	Objects    map[string]map[string]RenderContextEnvAwareObject
	Env        string
	Branch     string
	Namespace  string
	TagVersion string
}

func InjectMetadata(injectContext InjectContext, objects []map[string]interface{}) []map[string]interface{} {

	for _, object := range objects {
		kind := object["kind"].(string)

		var metadata = object["metadata"].(map[interface{}]interface{})
		name := metadata["name"].(string)
		envAwareObjectName := injectContext.Objects[kind][name].Name

		switch kind {
		case K8S_SERVICE:
			injectMetaDataIntoService(envAwareObjectName, injectContext, object)
			break
		case K8S_DEPLOYMENT:
			injectMetaDataIntoDeployment(envAwareObjectName, injectContext, object)
			break
		default:
			injectMetaDataIntoObject(envAwareObjectName, injectContext, object)
			break
		}
	}

	return objects
}

func injectMetaDataIntoService(envAwareObjectName string, injectContext InjectContext, object map[string]interface{}) map[string]interface{} {
	object = inject(envAwareObjectName, injectContext, object)

	var spec = object["spec"].(map[interface{}]interface{})

	var specSelector = spec["selector"].(map[interface{}]interface{})
	specSelector["env"] = injectContext.Env

	spec["selector"] = specSelector

	object["spec"] = spec

	return object
}

func injectMetaDataIntoDeployment(envAwareObjectName string, injectContext InjectContext, object map[string]interface{}) map[string]interface{} {
	object = inject(envAwareObjectName, injectContext, object)

	/*
	 * .spec.selector is an optional field that specifies a label selector for
	 * the Pods targeted by this deployment.
	 *
	 * If specified, .spec.selector must match .spec.template.metadata.labels,
	 * or it will be rejected by the API. If .spec.selector is unspecified,
	 * .spec.selector.matchLabels will be defaulted to .spec.template.metadata.labels.
	 *
	 * http://kubernetes.io/docs/user-guide/deployments/#selector
	 */
	spec := object["spec"].(map[interface{}]interface{})
	specTemplate := spec["template"].(map[interface{}]interface{})

	specTemplateMetadata := specTemplate["metadata"].(map[interface{}]interface{})

	var specTemplateMetadataLabels map[interface{}]interface{}

	if specTemplateMetadata["labels"].(map[interface{}]interface{}) == nil {
		specTemplateMetadataLabels = map[interface{}]interface{}{
			"env": injectContext.Env,
		}
	} else {
		specTemplateMetadataLabels = specTemplateMetadata["labels"].(map[interface{}]interface{})
		specTemplateMetadataLabels["env"] = injectContext.Env
	}

	if spec["selector"] == nil {
		spec["selector"] = map[interface{}]interface{}{
			"matchLabels": specTemplateMetadataLabels,
		}
	}

	specSelector := spec["selector"].(map[interface{}]interface{})

	/*
	 * Limit replica set history if not set.
	 */
	if spec["revisionHistoryLimit"] == nil {
		spec["revisionHistoryLimit"] = DEFAULT_REVISION_HISTORY_LIMIT
	}

	specSelectorMatchLabels := specSelector["matchLabels"].(map[interface{}]interface{})
	specSelectorMatchLabels["env"] = injectContext.Env

	specSelector["matchLabels"] = specSelectorMatchLabels

	spec["selector"] = specSelector

	object["spec"] = spec

	return object
}

func injectMetaDataIntoObject(envAwareObjectName string, injectContext InjectContext, object map[string]interface{}) map[string]interface{} {
	return inject(envAwareObjectName, injectContext, object)
}

func inject(envAwareObjectName string, injectContext InjectContext, object map[string]interface{}) map[string]interface{} {
	var metadata = object["metadata"].(map[interface{}]interface{})

	metadata["name"] = envAwareObjectName
	metadata["namespace"] = injectContext.Namespace

	if metadata["labels"] == nil {
		metadata["labels"] = map[interface{}]interface{}{}
	}

	var labels = metadata["labels"].(map[interface{}]interface{})

	labels["env"] = injectContext.Env
	labels["branch"] = injectContext.Branch
	labels["branch_hash"] = MD5(injectContext.Branch)

	labels["version"] = injectContext.TagVersion

	metadata["labels"] = labels
	object["metadata"] = metadata

	return object
}

func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
