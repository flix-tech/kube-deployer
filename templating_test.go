package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBuildRenderContext(t *testing.T) {

	var deployerSpec = DeployerSpec{
		ProjectDir: ".",
		Cluster: DeployerSpecCluster{
			Host:  "localhost",
			Token: "secret",
		},
		TagVersion: "test",
		Env:        "master",
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

	if err != nil {
		t.Error(err.Error())
	}

	var expected = "master-test-service"
	var actual = renderContext.Objects["Service"]["test-service"].Name

	assert := assert.New(t)

	assert.Equal(expected, actual, "The two words should be the same.")

	expected = "master-test-deployment"
	actual = renderContext.Objects["Deployment"]["test-deployment"].Name

	assert.Equal(expected, actual, "The two words should be the same.")
}
