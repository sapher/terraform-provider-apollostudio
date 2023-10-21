---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "apollostudio Provider"
subcategory: ""
description: |-
  Interact with Apollo Studio
---

# apollostudio Provider

Interact with Apollo Studio

## Example Usage

```terraform
# Simple example using default endpoint and api key taken from APOLLO_KEY env variable
provider "apollostudio" {
  org_id = "your-org-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `org_id` (String) Organization ID on Apollo Studio. Can also be set via the APOLLO_ORG_ID environment variable

### Optional

- `api_key` (String) API key for the Apollo GraphQL API. Can also be set via the APOLLO_KEY environment variable
- `host` (String) Host of the Apollo GraphQL API. Defaults to https://graphql.api.apollographql.com/api/graphql