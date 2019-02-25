package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var __VERSION__ string

const K8S_SERVICE = "Service"
const K8S_DEPLOYMENT = "Deployment"
const K8S_CRONJOB = "CronJob"

const K8S_API_VERSION_BATCH_V2_ALPHA1 = "batch/v2alpha1"

const PROJECT_DIR_FLAG = "project-dir"
const TAG_FLAG = "tag"
const CLUSTER_FLAG = "cluster"
const NAMESPACE_FLAG = "namespace"
const ENV_FLAG = "env"
const BRANCH_FLAG = "branch"
const TEMPLATE_FLAG = "template"
const CONTAINER_FLAG = "container"
const SERVER_FLAG = "server"
const DRY_RUN_FLAG = "dry-run"
const VERBOSE_FLAG = "verbose"
const TOKEN_FLAG = "token"
const CONTEXT_FLAG = "context"

var projectDirFlag = cli.StringFlag{
	Name:  PROJECT_DIR_FLAG,
	Usage: "Path where .kube-deploy.yml lives",
}
var tagFlag = cli.StringFlag{
	Name:  TAG_FLAG,
	Usage: "Container tag.",
}
var clusterFlag = cli.StringFlag{
	Name:  CLUSTER_FLAG,
	Usage: "Cluster name.",
}
var namespaceFlag = cli.StringFlag{
	Name:  NAMESPACE_FLAG,
	Usage: "Kubernetes namespace.",
}
var envFlag = cli.StringFlag{
	Name:  ENV_FLAG,
	Usage: "Kubernetes env. Optional. Will use branch value if omitted. Example branch=master, env=production, branch=master, env=staging.",
}
var branchFlag = cli.StringFlag{
	Name:  BRANCH_FLAG,
	Usage: "Git branch. Will be injected to all kube objects labels. Also a branch_hash label will be injected which will be used to clean resources of merged branches.",
}
var templateFlag = cli.StringSliceFlag{
	Name:  TEMPLATE_FLAG,
	Usage: "Template file. Only provide this if you are using the non config file mode.",
}
var containerFlag = cli.StringSliceFlag{
	Name:  CONTAINER_FLAG,
	Usage: "Container tag in format container_id:container_image_name. Only provide this if you are using the non config file mode.",
}
var serverFlag = cli.StringFlag{
	Name:  SERVER_FLAG,
	Usage: "Kube server address. Only provide this if you are using the non config file mode.",
}
var dryRunFlag = cli.BoolFlag{
	Name:  DRY_RUN_FLAG,
	Usage: "Use kubectl --dry-run",
}
var verboseFlag = cli.BoolFlag{
	Name:  VERBOSE_FLAG,
	Usage: "Display more",
}
var tokenFlag = cli.StringFlag{
	Name:  TOKEN_FLAG,
	Usage: "Kube token. Alternative it will try to read KUBE_TOKEN env variable.",
}
var contextFlag = cli.StringFlag{
	Name:  CONTEXT_FLAG,
	Usage: "Kube context. Will switch to this context before runnig any kubectl command",
}

func main() {
	app := cli.NewApp()
	app.Version = version()
	app.Description = "Render & deploy kubernetes definitions"
	app.Name = "kube-deployer"
	app.Commands = []cli.Command{
		{
			Name: "deploy",
			Flags: []cli.Flag{
				projectDirFlag,
				tagFlag,
				clusterFlag,
				namespaceFlag,
				envFlag,
				branchFlag,
				templateFlag,
				containerFlag,
				serverFlag,
				dryRunFlag,
				verboseFlag,
				tokenFlag,
				contextFlag,
			},
			Action: func(c *cli.Context) error {
				dryRun := c.Bool(DRY_RUN_FLAG)
				verbose := c.Bool(VERBOSE_FLAG)
				token := c.String(TOKEN_FLAG)
				context := c.String(CONTEXT_FLAG)

				var clusterApiToken string
				if token == "" {
					clusterApiToken = os.Getenv("KUBE_TOKEN")
				} else {
					clusterApiToken = token
				}

				if clusterApiToken == "" && context == "" {
					log.Fatal("Please provide a Kubernetes access token or context.")
				}

				var deployerSpec DeployerSpec
				err := deployerSpec.FromCliContext(c)

				if err != nil {
					log.Fatalf("error: %v", err)
				}
				deployerSpec.Cluster.Token = clusterApiToken
				deployerSpec.Cluster.Context = context

				deploy(deployerSpec, dryRun, verbose)

				return nil
			},
		},
		{
			Name: "render",
			Flags: []cli.Flag{
				projectDirFlag,
				tagFlag,
				clusterFlag,
				namespaceFlag,
				envFlag,
				branchFlag,
				templateFlag,
				containerFlag,
				serverFlag,
			},
			Action: func(c *cli.Context) error {
				var deployerSpec DeployerSpec
				err := deployerSpec.FromCliContext(c)

				if err != nil {
					log.Fatalf("error: %v", err)
				}

				kubernetesDefinition, err := render(deployerSpec)

				if err != nil {
					log.Fatalf("error: %v", err)
				}

				fmt.Println(kubernetesDefinition)

				return nil
			},
		},
		{
			Name: "clean",
			Flags: []cli.Flag{
				projectDirFlag,
				clusterFlag,
				namespaceFlag,
				serverFlag,
				tokenFlag,
				contextFlag,
			},
			Action: func(c *cli.Context) error {
				projectDir := c.String(PROJECT_DIR_FLAG)
				cluster := c.String(CLUSTER_FLAG)
				token := c.String(TOKEN_FLAG)
				context := c.String(CONTEXT_FLAG)

				if projectDir == "" {
					projectDir = "."
				}

				var server = c.String(SERVER_FLAG)
				var namespace = c.String(NAMESPACE_FLAG)

				if cluster == "" && server == "" {
					fmt.Println("Please specify either the cluster or server flag.")
					os.Exit(1)
				}

				if cluster != "" {
					var deployerConfigFile DeployerConfigFile
					err := deployerConfigFile.ReadFileFromFile(projectDir)

					if err != nil {
						log.Fatalf("error: %v", err)
					}

					server = deployerConfigFile.Clusters[cluster].Host
				}

				var clusterApiToken string
				if token == "" {
					clusterApiToken = os.Getenv("KUBE_TOKEN")
				} else {
					clusterApiToken = token
				}

				if clusterApiToken == "" && context == "" {
					log.Fatal("Please provide a Kubernetes access token or context.")
				}

				clean(projectDir, server, clusterApiToken, namespace, context)

				return nil
			},
		},
	}

	app.Run(os.Args)
}

func clean(projectDir string, host string, token string, namespace string, context string) {
	kubectl := KubeClient{
		Token:   token,
		Server:  host,
		Verbose: false,
		Context: context,
	}

	deployedBranchHashes, err := kubectl.GetDeployedBranchHashes(namespace)

	if err != nil {
		log.Fatal(err)
	}

	projectBranchHashes := BranchHashes(projectDir)
	branchesHashesToDelete := []string{}

	for _, branchHash := range deployedBranchHashes {
		if _, ok := projectBranchHashes[branchHash]; !ok {
			branchesHashesToDelete = append(branchesHashesToDelete, branchHash)
		}
	}

	for _, branchHashToDelete := range branchesHashesToDelete {
		output, err := kubectl.DeleteObjectsByBranch(branchHashToDelete, namespace)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(output)
	}
}

func deploy(deployerSpec DeployerSpec, dryRun bool, verbose bool) {
	kubernetesDefinition, err := render(deployerSpec)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var kubeCtl = KubeClient{
		Token:   deployerSpec.Cluster.Token,
		Server:  deployerSpec.Cluster.Host,
		Context: deployerSpec.Cluster.Context,
		Verbose: verbose,
	}

	kubeCtl.Version()

	kubeCtl.Apply(kubernetesDefinition, deployerSpec.Namespace, dryRun)
}

func render(deployerSpec DeployerSpec) (string, error) {
	objects, err := deployerSpec.ParseKubernetesYamlFiles()

	if err != nil {
		log.Fatal(err.Error())
	}

	var renderContext RenderContext
	err = renderContext.Build(deployerSpec, objects)

	if err != nil {
		log.Fatal(err.Error())
	}

	injectContext := InjectContext{
		Objects:    renderContext.Objects,
		Env:        renderContext.Env,
		Branch:     renderContext.Branch,
		Namespace:  renderContext.Namespace,
		TagVersion: renderContext.DeployerSpec.TagVersion,
	}
	objects = InjectMetadata(injectContext, objects)

	templates := make([]string, 0)

	for _, object := range objects {
		template, err := yaml.Marshal(object)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		templates = append(templates, string(template))
	}

	return renderContext.Render(templates)
}

func version() string {
	if __VERSION__ == "" {
		return "dev"
	}

	return __VERSION__
}
