# sortKeyHttpServer

> This project scrapes statistics for websites provided in JSON format. It then processes and returns the data in accordance with user-provided inputs.

## Table of contents

- [Usage](#usage)
- [Getting started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installing and running](#installing-and-running)
    - [Installing and Running - Locally](#installing-and-running---locally)
    - [Installing and Running - Go binary](#installing-and-running-go---binary)
- [Running the tests](#running-the-tests)
  - [End to end tests](#end-to-end-test)
  - [Coding style tests](#coding-style-tests)
- [Deployment](#deployment)
  - [Deployment using Docker Compose](#deployment-using-docker-compose)
  - [K8s Deployment using yaml manifests](#k8s-deployment-using-yaml-manifests)
  - [K8s Deployment using Kustomize](#k8s-deployment-using-kustomize)
- [Contributing](#contributing)

## Getting Started

The application can collect the URL statistics data from a file or a series of HTTP endpoints. The URL statistics information is provided in a JSON format.
The Data Collection Method and the Data Collection Source can be overridden using the Environment Variables `DATA_COLLECTION_METHOD` and `DATA_COLLECTION_PATH`.
The default Data Collection Method is `http`, but it can be overridden to `file`.

After deploying the application, it will be available for access at localhost in either port 5000 or 80 (depending on the deployment method).
The services are provided over the following URL's:
- Raw data: `http://localhost/`
- Sorted by Relevance Score: `http://localhost/sortkey/relevanceScore`
- Sorted by View: `http://localhost/sortkey/views`

The optional parameter `limit` can be used to limit the return. The optional parameter is only applicable to the Sorted URL's
- Limit Sorted by Relevance Score: `http://localhost/sortkey/relevanceScore?limit=3`
- Limit Sorted by Views: `http://localhost/sortkey/views?limit=5`

A full list of instructions can be obtained by running `make help` in the root directory:

```
$make help

Usage:
  make <target>

General
  help             Display this help.

Development
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  run              Run the webservice from host.

Build
  build            Build go binary.
  docker-build     Build docker image with the webservice.
  kind-load-docker-image  Push docker image with the webservice to kind cluster.

Deployment Infra
  kind-create-cluster  Deploy Kind k8s cluster.
  kind-delete-cluster  Delete Kind k8s cluster.

Deployment App
  deploy-kustomize  Deploy urlstats app to the K8s cluster specified in ~/.kube/config using Kustomize.
  undeploy-kustomize  Undeploy urlstats app to the K8s cluster specified in ~/.kube/config using Kustomize.
  deploy-k8s       Deploy urlstats app to the K8s cluster specified in ~/.kube/config using yaml files.
  undeploy-k8s     Undeploy urlstats app to the K8s cluster specified in ~/.kube/config using yaml files.
  deploy-docker-file  Deploy urlstats app locally using docker.
  deploy-docker-http  Deploy urlstats app locally using docker.
  deploy-bin       Deploy urlstats app locally using go binary.

E2E Deployment
  all              build docker image, provision k8's kinD cluster and deploy urlstats webservice using kustomize

Support/Troubleshoot
  kind-list-loaded-images  List Docker images loaded to Kind k8s cluster.
  wsl2-start-docker-daemon  Start docker daemon in wsl2.

Cleanup
  clean-bin        Removes the bin directory
  clean-docker-images  Removes all docker images
  clean-docker-containers  Removes all docker containers. WARNING: also removes running containers.

Build Dependencies
  --kind           Download kind locally if necessary.
  --kustomize      Download kustomize locally if necessary.
```
Quick deployment can be achieved by running `make all`, this creates a KinD cluster then deploys and configures the application using Kustomize.
See [Deployment](#deployment) section for more deployment options.
## Prerequisites

This project relies solely on the go standard library.
For deployment purposes, Docker must be installed. Install from [here](https://www.docker.com/products/docker-desktop/).

Installation of Kustomize and KinD are managed automatically with the Makefile and their binaries are used when necessary. The binaries for managed third-party applications can be found in the `bin/` folder.

## Installing and Running

A number of options have been provided for installing and running the application.
They are all accessible in the Makefile.

### Installing and Running - Locally

In the root directory:

```sh
make run
```

Access the application using port 5000:
- `http://localhost:5000/`

### Installing and Running - Go binary

In the root directory:

```sh
make build
make deploy-bin
```

Access the application using port 5000:
- `http://localhost:5000/`

## Running the tests

For testing purposes, it is preferred to run the application using the `file` Data Collection Method.

### Running tests from terminal

The recommendation is to use the Makefile. In the root directory:

```sh
make test-unit
```

### Running/Debugging tests from Visual Studio Code

Add the following section to your `launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "env": {
                "DATA_COLLECTION_METHOD": "file",
                "DATA_COLLECTION_PATH": "dev-resources/raw-json-files"
            }
        }
    ]
}
```

## Deployment

A number of deployment options have been provided.
They are all accessible in the Makefile.

### Deployment using Docker Compose

Using Data Collection Method: HTTP

```sh
make docker-build
make deploy-docker-http
```

Using Data Collection Method: FILE

```sh
make docker-build
make deploy-docker-file
```

The application can be accessed using port 80:
- `http://localhost/`

The application will bind to the terminal and provide live logs.
Type `Ctrl+C` to stop the docker deployment.

### K8s Deployment using yaml manifests

We have chosen to use KinD clusters for local testing

```sh
make docker-build
make kind-create-cluster
make kind-load-docker-image
make deploy-k8s
```
The application can be accessed using port 80:
- `http://localhost/`

To clean-up, simply undeploy the application or remove the KinD Cluster
```sh
make undeploys-k8s
make kind-delete-cluster
```

### K8s Deployment using Kustomize

We have chosen to use KinD clusters for local testing

```sh
make docker-build
make kind-create-cluster
make kind-load-docker-image
make deploy-kustomize
```
The application can be accessed using port 80:
- `http://localhost/`

To clean-up, simply undeploy the application or remove the KinD Cluster
```sh
make undeploys-kustomize
make kind-delete-cluster
```

## Contributing

1. Fork it (<https://github.com/felipe88alves/sortKeyHttpServer/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

---

<!-- Markdown link & img definitions -->
