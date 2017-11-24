package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInjectMetadata(t *testing.T) {
	var deployerSpec = DeployerSpec{
		ProjectDir: ".",
		Cluster: DeployerSpecCluster{
			Host:  "localhost",
			Token: "secret",
		},
		TagVersion: "test-01",
		Env:        "static-staging",
		Branch:     "master",
		Namespace:  "default",
		Templates:  []string{"./k8s/app.yml"},
		Containers: []DeployerSpecContainer{
			{
				Image: "test",
				Id:    "test",
			},
		},
	}

	yaml := `
kind: Service
apiVersion: v1
metadata:
  name: test-service
spec:
  ports:
    -
      name: http
      protocol: TCP
      port: 80
  selector:
    app: test-service

---

apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test-deployment
spec:
  template:
    metadata:
      labels:
        app: test-deployment
    spec:
      containers:
        - name: debian
          image: debian:latest

`

	objects, err := UnmarshalYaml(yaml)

	if err != nil {
		t.Error(err.Error())
	}

	var renderContext RenderContext
	err = renderContext.Build(deployerSpec, objects)

	injectContext := InjectContext{
		Objects:    renderContext.Objects,
		Env:        renderContext.Env,
		Branch:     renderContext.Branch,
		Namespace:  renderContext.Namespace,
		TagVersion: renderContext.DeployerSpec.TagVersion,
	}
	objects = InjectMetadata(injectContext, objects)

	assert := assert.New(t)

	expectedNamespace := "default"
	expectedVersion := "test-01"
	expectedEnv := "static-staging"
	expectedBranch := "master"
	exptectedBranchHash := "eb0a191797624dd3a48fa681d3061212"

	for _, object := range objects {
		metadata := object["metadata"].(map[interface{}]interface{})
		kind := object["kind"].(string)

		labels := metadata["labels"].(map[interface{}]interface{})
		actualEnv := labels["env"].(string)
		actualBranch := labels["branch"].(string)
		actualBranchHash := labels["branch_hash"].(string)
		actualVersion := labels["version"].(string)

		actualName := metadata["name"].(string)
		actualNamespace := metadata["namespace"].(string)

		assert.Equal(expectedNamespace, actualNamespace)
		assert.Equal(expectedEnv, actualEnv)
		assert.Equal(expectedVersion, actualVersion)
		assert.Equal(expectedBranch, actualBranch)
		assert.Equal(exptectedBranchHash, actualBranchHash)

		var expectedName string

		switch kind {
		case "Service":
			expectedName = "static-staging-test-service"

			expectedSpecSelectorEnv := "static-staging"

			var spec = object["spec"].(map[interface{}]interface{})
			var specSelector = spec["selector"].(map[interface{}]interface{})
			actualSpecSelectorEnv := specSelector["env"].(string)

			assert.Equal(expectedSpecSelectorEnv, actualSpecSelectorEnv)
			break
		case "Deployment":
			expectedName = "static-staging-test-deployment"

			break
		default:
			break
		}

		assert.Equal(expectedName, actualName)
	}
}
