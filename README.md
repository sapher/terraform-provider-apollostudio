# ApolloStudio Terraform Provider

Management of Apollo Studio resources like graphs, graph variants, graph api keys using Terraform.
Offers many data sources to query the Apollo Studio API.

:warning: This provider is in early development and is not yet ready for production use. Works needs to be done to improve the documentation, testing, and error handling.

Any contributions are welcome!

## Local development

**Clone repository**

```bash
git clone git@github.com:sapher/terraform-provider-apollostudio.git ~/terraform-provider-apollostudio
```

Be mindful that should keep the destination folder name as `terraform-provider-apollostudio`, otherwise the content of the `docs` folder will be rendered using a different prefix than the one awaited.

**Configure local development**

Next, you need to create or update the file `~/.terraformrc` with the following content:

```hcl
provider_installation {
  dev_overrides {
    "hashicorp.com/sapher/apollostudio" = "/home/<your-username>/terraform-provider-apollo"
  }

  direct {}
}
```

Note that the path to the provider is absolute and should point to the folder where you cloned the repository.

This will force terraform to use the local provider instead of the one downloaded from the registry.

**Development usage**

Then you can use it in your code like so :

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

> You don't have to run `terraform init` as the provider is run from the local machine.

> Note that many fields are missing from resource and data sources. It's on purpose as depending of the user roles, many fields are not available. Feel free to request the addition of a field if you need it.

## Links

- [Terraform Registry](https://registry.terraform.io/providers/sapher/apollostudio/latest)
- [GraphOS Platform API Documentation](https://www.apollographql.com/docs/graphos/platform-api/)
