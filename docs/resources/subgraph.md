---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "apollostudio_subgraph Resource - terraform-provider-apollostudio"
subcategory: ""
description: |-
  Manage a subgraph
---

# apollostudio_subgraph (Resource)

Manage a subgraph

## Example Usage

```terraform
resource "apollostudio_subgraph" "this" {
  graph_id     = "your-graph-id"
  variant_name = "your-variant-name"
  name         = "your-subgraph-name"
  url          = "your-graphql-endpoint-url"
  schema       = file("path-to-your-schema.graphql")
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `graph_id` (String) ID of the graph
- `name` (String) Name of the subgraph
- `schema` (String) Schema of the subgraph variant
- `url` (String) URL of the subgraph variant
- `variant_name` (String) Name of the subgraph variant

### Read-Only

- `revision` (String) Revision of the subgraph variant

## Import

Import is supported using the following syntax:

```shell
terraform import apollostudio_subgraph.example your-graph-id@variant-name:subgraph-name
```
