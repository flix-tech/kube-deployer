package main

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

const DEFAULT_DEPLOYER_YAML = ".kube-deploy.yml"
const DEPLOYER_SPEC_MIN_VERSION = 1

func (spec *DeployerSpec) FromCliContext(c *cli.Context) error {
	tag := c.String(TAG_FLAG)
	namespace := c.String(NAMESPACE_FLAG)
	env := c.String(ENV_FLAG)
	branch := c.String(BRANCH_FLAG)
	projectDir := c.String(PROJECT_DIR_FLAG)

	// config-file specific parameters
	cluster := c.String(CLUSTER_FLAG)

	// Non config-file parameters
	templates := c.StringSlice(TEMPLATE_FLAG)
	containers := c.StringSlice(CONTAINER_FLAG)
	server := c.String(SERVER_FLAG)

	configMode := len(templates) == 0 && len(containers) == 0 && server == ""

	requiredFlags := map[string]interface{}{
		"tag":       tag,
		"namespace": namespace,
		"branch":    branch,
		"env":       env,
	}

	if configMode {
		requiredFlags["cluster"] = cluster
	} else {
		requiredFlags["template"] = templates
		requiredFlags["containers"] = containers
		requiredFlags["server"] = server
	}

	if env == "" {
		env = branch
	}

	for k, v := range requiredFlags {
		if v == "" || v == nil {
			fmt.Println("Please specify the " + k + " flag.")
			os.Exit(1)
		}
	}

	if projectDir == "" {
		projectDir = "."
	}

	var err error
	if configMode {
		err = spec.fromFile(
			projectDir,
			tag,
			cluster,
			env,
			namespace,
		)
	} else {
		err = spec.fromParameters(
			projectDir,
			server,
			tag,
			env,
			namespace,
			templates,
			containers,
		)
	}

	spec.Branch = branch

	return err
}

func (deployerConfig *DeployerConfigFile) ReadFileFromFile(projectDir string) error {
	filePath := projectDir + "/" + DEFAULT_DEPLOYER_YAML
	yamlFile, err := ioutil.ReadFile(filePath)

	if err != nil {
		return errors.New("Cannot read file " + filePath)
	}

	err = yaml.Unmarshal(yamlFile, deployerConfig)

	if err != nil {
		return err
	}

	if deployerConfig.SpecVersion < DEPLOYER_SPEC_MIN_VERSION {
		return errors.New(".kube-deploy.yml Spec must have at least version 3")
	}

	return nil
}

func (spec *DeployerSpec) fromFile(
	projectDir string,
	tag string,
	cluster string,
	env string,
	namespace string,
) error {

	var deployerConfig DeployerConfigFile
	err := deployerConfig.ReadFileFromFile(projectDir)

	if err != nil {
		return err
	}

	spec.ProjectDir = projectDir
	spec.TagVersion = tag
	spec.Namespace = namespace
	spec.Env = env

	clusterDefinition, clusterExist := deployerConfig.Clusters[cluster]

	if !clusterExist {
		return errors.New(fmt.Sprintf("cluster %s not present in list of clusters", cluster))
	}

	spec.Cluster = DeployerSpecCluster{
		Host: clusterDefinition.Host,
	}

	spec.Containers = make([]DeployerSpecContainer, len(deployerConfig.Containers))
	for _, container := range deployerConfig.Containers {
		spec.Containers = append(spec.Containers, DeployerSpecContainer{
			Image: container.Image,
			Id:    container.Id,
		})
	}

	namespaceExist := false
	for _, target := range clusterDefinition.Targets {
		if target.Namespace == namespace {
			namespaceExist = true

			spec.Templates = target.Templates
		}
	}

	if !namespaceExist {
		return errors.New(fmt.Sprintf("Namespace %s not present in list of targets", namespace))
	}

	return nil
}

func (spec *DeployerSpec) fromParameters(
	projectDir string,
	server string,
	tag string,
	env string,
	namespace string,
	templates []string,
	containers []string,
) error {

	spec.ProjectDir = projectDir
	spec.TagVersion = tag
	spec.Namespace = namespace
	spec.Env = env
	spec.Templates = templates

	spec.Cluster = DeployerSpecCluster{
		Host: server,
	}

	spec.Containers = make([]DeployerSpecContainer, len(containers))
	for _, container := range containers {
		containerParts := strings.Split(container, ":")

		spec.Containers = append(spec.Containers, DeployerSpecContainer{
			Id:    containerParts[0],
			Image: containerParts[1],
		})
	}

	return nil
}

func (spec *DeployerSpec) ParseKubernetesYamlFiles() ([]map[string]interface{}, error) {
	var objects = make([]map[string]interface{}, 0)

	for _, template := range spec.Templates {
		filePath := spec.ProjectDir + "/" + template

		yamlFile, err := ioutil.ReadFile(filePath)

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot read file %s", filePath))
		}

		unmarshaledObjects, err := UnmarshalYaml(string(yamlFile))

		if err != nil {
			return nil, err
		}

		for _, object := range unmarshaledObjects {
			objects = append(objects, object)
		}
	}

	return objects, nil
}

type DeployerSpec struct {
	ProjectDir string
	TagVersion string
	Env        string
	Branch     string
	Namespace  string
	Containers []DeployerSpecContainer
	Templates  []string
	Cluster    DeployerSpecCluster
}

type DeployerSpecCluster struct {
	Host    string
	Token   string
	Context string
}

type DeployerSpecContainer struct {
	Id    string
	Image string
}

type DeployerConfigFile struct {
	SpecVersion int `yaml:"version"`
	Containers  []struct {
		Id    string `yaml:"id"`
		Image string `yaml:"image"`
	} `yaml:"containers"`
	Clusters map[string]struct {
		Host    string                     `yaml:"host"`
		Targets []DeployerConfigFileTarget `yaml:"targets"`
	} `yaml:"clusters"`
}

type DeployerConfigFileTarget struct {
	Namespace string   `yaml:"namespace"`
	Templates []string `yaml:"templates"`
}
