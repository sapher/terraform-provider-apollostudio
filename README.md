# Apollo GraphQL Terraform Provider

![GitHub Release](https://img.shields.io/github/v/release/sapher/terraform-provider-apollostudio)
![Downloads](https://img.shields.io/github/downloads/sapher/terraform-provider-apollostudio/total)

Unnoficial Terraform provider for [Apollo GraphQL][Apollo]. Enable management of resources like graphs, graph variants, subgraphs, graph API keys, and more using Terraform.

The killer feature of this provider is the extensive [schema checks](https://www.apollographql.com/docs/graphos/delivery/schema-checks) done before a subgraph is created or updated. This make it virtually impossible to deploy an invalid subgraph and break the supergraph because of it. If an errors or warnings are detected, the provider will nicely display them in the Terraform output.

## Usage example

In order to use this provider you will need to add the following code to your Terraform configuration file:

```hcl
terraform {
  required_providers {
    apollostudio = {
      source = "sapher/apollostudio"
      version = ">= 1.3.0"
    }
  }
}

provider "apollostudio" {
  api_key = "<your-api-key>"
  org_id  = "<your-org-id>"
}
```

All the resources and data sources are documented in the [provider documentation page](https://registry.terraform.io/providers/sapher/apollostudio/latest/docs).

## Contributions

Any contributions are welcome!

## Local development

In the following section, we will see how to setup a local development environment for the provider. This will allow you to test your changes locally and ease the development process.

### Prerequisites

**API Key**

First you will need an account on [Apollo GraphQL][Apollo] and you will need to generate an [API Key](https://www.apollographql.com/docs/graphos/api-keys/) with the `GRAPH_ADMIN` role for the whole organization. It is recommended to use API key with this role so that all features of the provider are available. Otherwise, if you have to use another role, some resources won't be available.

**Tools**

The following tools need to be installed on your machine and available in your `PATH`:

- [golang](https://golang.org/doc/install)
- [terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)

The following tools are optional but highly recommended. They are used to run checks and tests locally:

- [pre-commit](https://pre-commit.com/#install) : Run checks before committing code
- [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) : Linter for Golang codebase
- [gitleaks](https://github.com/gitleaks/gitleaks) : Analyze git history to find secrets

### Clone repository

```bash
git clone git@github.com:sapher/terraform-provider-apollostudio.git $HOME/terraform-provider-apollostudio
```

This will clone this repository in the `$HOME/terraform-provider-apollostudio` folder. It is important to keep the folder name as `terraform-provider-apollostudio` otherwise the content of the `docs` folder will be rendered by the `make generate` command using a different prefix than the one awaited.

### Configure local environment

Next, you need to create or update the file `~/.terraformrc` with the following content:

```hcl
provider_installation {
  dev_overrides {
    "hashicorp.com/sapher/apollostudio" = "/home/<your-username>/terraform-provider-apollo"
  }

  direct {}
}
```

This will force terraform to use the local provider instead of the published one from the registry.

Note that the path needs to be absolute and should point to the folder where you cloned the repository.

Now you need to build the provider:

```bash
cd $HOME/terraform-provider-apollostudio
make build
```

An executable file named `terraform-provider-apollostudio` should have been created in the repository root folder.

### Usage of the development provider

Now you can create a new `main.tf` file with the following content:

```hcl
terraform {
  required_providers {
    apollostudio = {
      source = "hashicorp.com/sapher/apollostudio"
    }
  }
}

provider "apollostudio" {
  api_key = "<your-api-key>"
  org_id  = "<your-org-id>"
}

data "apollostudio_graphs" "this" {}
```

Unless you use other providers in your code, you don't have to run `terraform init` as the provider is run from the local machine.

### Install git hooks

We use pre-commit to run some checks before committing any code. To install the git hooks, run the following command:

```bash
pre-commit install
```

Be sure to have the tools listed in the [prerequisites](#prerequisites) section installed and available in your `PATH`.

You can also run pre-commit at any time by running the following command:

```bash
pre-commit run --all-files
```

### Runs acceptance tests locally

Before running any of the following commands, you need to export these environment variables:

```bash
export APOLLO_API_KEY=<your-api-key>
export APOLLO_ORG_ID=<your-org-id>
```

As acceptance tests are run against a real Apollo GraphQL organization, we need to create a dedicated graph for the tests.

For this you need to go to your dashboard and create a new graph with the following settings:

| Name          | Value                                                       |
| ------------- | ----------------------------------------------------------- |
| Graph Name    | `testacc-terraform`                                         |
| Graph ID      | `testacc-terraform` - :warning: do not use autogenerated id |
| Description   | `graph dedicated for terraform provider acceptance testing` |
| Subgraph Name | `countries`                                                 |
| Routing URL   | `https://countries.trevorblades.com`                        |

Then you can run the acceptance tests :

```bash
make testacc
```

## Disclaimer

> This project is not affiliated in any with Apollo Graph Inc. All registered trademarks are the property of their respective owners.

[Apollo]: https://www.apollographql.com/
