![](https://travis-ci.org/flix-tech/kube-deployer.svg?branch=master)


Kube Deployer adds templating and multi-branch environments to your K8S deployments, 
reduces your project overhead and takes care of cleaning up merged branches.

If you want templating in your Kubernetes files and/or multi branch support please continue reading.

How to use

Go see all available commands please run:

```
$: kube-deploy -h
```

## Templating

Internally kube-deployer uses the handlebars v3 templating engine.

Inside the template you have access to a "context" variable of type RenderContext.

```
type RenderContext struct {
    Containers   map[string]RenderContextContainer
    Objects      map[string]map[string]RenderContextEnvAwareObject
    Env          string // url slugged
    Namespace    string
    DeployerSpec DeployerSpec
}

type RenderContextContainer struct {
    Name  string
}

type RenderContextEnvAwareObject struct {
    Name string
}

type DeployerSpec struct {
    ProjectDir string
    TagVersion string
    Env        string
    Namespace  string
    Containers []DeployerSpecContainer
    Templates  []string
    Cluster    DeployerSpecCluster
}

type DeployerSpecCluster struct {
    Host  string
    Token string
}

type DeployerSpecContainer struct {
    Id    string
    Image string
}

```

If you name a container "php" as id, then you reference it in your template like this:

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
    name: foo-web
    labels:
        app: foo-web
spec:
    replicas: 1
    template:
        metadata:
        labels:
            app: foo-web
        spec:
            containers:
                - name: php
                  image: "{{ context.containers.php.name }}"
...
```

# Multi env deployments

kube-deployer is changing the metadata.name of all objects to include the env you passed.


To get the final object name you can access the context.objects map. (Templating is case sensitive)

```
{{ context.objects.{kind}.{original-metadata-name}.name }}

// kind must be a valid k8s object (Service|Deployment|...)

Example:

volumes:
  - name: db-volume
    persistentVolumeClaim:
      claimName: '{{ context.objects.PersistentVolumeClaim.foo-pvc.name }}'
```

   
## Provide a Kube Access Token

Either set the KUBE_TOKEN env variable or pass the token via the -token=xxx flag.
   
## Dry Run & Verbose

To debug locally you can pass the --dry-run and --verbose flag.

```
$: kube-deploy deploy ... --dry-run --verbose
```

## No config file mode

In the "no-config-file" mode you need to pass all information like api host, templates, containers as cli arguments.

Following arguments are requiered:
* template (string list)
* containers (string list in format <id>:<registry/image> the id is a key chosen by  you)
* server (api host of the cluster)

```
$: kube-deploy deploy \
    -env=production \
    -branch=master \
    -tag=1.4.5 \
    -namespace=staging-foo \
    -template=./kubernetes/staging/web.yml \
    -template=./kubernetes/staging/db.yml \
    -server=https://foo.k8s.io \
    -container=php:foo/bar
```

## Config file mode

You can place a .kube-deploy.yml file in your project and store multiple cluster data and targets and choose based on the -cluster and -namespace cli flag.
 
Example:

```    
version: 1

containers:
    - id: "php"
      image: "foo/bar"

clusters:
    de_cluster:
        host: https://foo.k8s.bar.io
        targets:
            - namespace: staging-foo
              templates:
                - "./kubernetes/staging/web.yml"
                - "./kubernetes/staging/db.yml"
                - "./kubernetes/staging/pvc.yml"
                - "./kubernetes/staging/consumer.yml"
            - namespace: prod-foo
              templates:
                - "./kubernetes/prod/web.yml"
                - "./kubernetes/prod/qa.yml"
                - "./kubernetes/prod/consumer.yml"
                - "./kubernetes/prod/crons.yml"
```

The -cluster flag is required when running with config file

```
$: kube-deploy deploy \
    -env=prod \
    -branch=master \
    -tag=1.4.5 \
    -namespace=staging-foo \
    -cluster=de_cluster
```
	
## Running outside of your project directory

When running kube-deploy outside of your project directory you will need to provide the absolute path via the -project-dir flag

```
$: kube-deploy deploy ... -project-dir=/user/app
```

## Only render the templates

If don't want to use the internal "kubectl apply" command from kube-deployer you can run the "render" command.

For help run: `kube-deploy render -h`

It takes the same arguments (without token, verbose, dry-run because they are not needed for rendering) which will work the same as the "deploy" command but it will output the rendered template on STDOUT.
    
This way you can do further manipulation or customize the kubectl apply call with your needs.

## Gitlab CI usage

Here is an example of how to multi-branch-deploy from gitlab-ci.

```
# deploy on branch pushes
deploy-staging-branch:
  image: flixtech/kube-deployer:latest
  script: >
    kube-deploy
    deploy
    -branch=$CI_BUILD_REF_NAME
    -env=$CI_BUILD_REF_NAME
    -tag=$CI_PIPELINE_ID
    -namespace=staging-foo
    -cluster=de_cluster
  stage: deploy
  only:
    - branches
  except:
    - master

# deploy latest master to static-1 and static-2 staging env
deploy-sandbox-branch:
  image: flixtech/kube-deployer:latest
  script:
	- kube-deploy deploy \
	    -branch=master \
	    -env=static-1 \
	    -tag=$CI_PIPELINE_ID \
	    -namespace=staging-foo \
	    -cluster=de_cluster
	- kube-deploy deploy \
	    -branch=master \
	    -env=static-2 \
	    -tag=$CI_PIPELINE_ID \
	    -namespace=staging-foo \
	    -cluster=de_cluster
  stage: deploy
  only:
    - master
```
