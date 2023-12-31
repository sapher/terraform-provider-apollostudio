---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "apollostudio_graph_api_keys Data Source - terraform-provider-apollostudio"
subcategory: ""
description: |-
  Provide details about a specific graph's API keys. Beware that the API key token is partially masked when read, it's only available at creation time.
---

# apollostudio_graph_api_keys (Data Source)

Provide details about a specific graph's API keys. Beware that the API key token is partially masked when read, it's only available at creation time.

## Example Usage

```terraform
data "apollostudio_graph_api_keys" "this" {
  graph_id = "your-graph-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `graph_id` (String) ID of the graph linked to the API keys

### Read-Only

- `api_keys` (Attributes List) List of API keys (see [below for nested schema](#nestedatt--api_keys))

<a id="nestedatt--api_keys"></a>
### Nested Schema for `api_keys`

Read-Only:

- `created_at` (String) Creation date of the API key
- `id` (String) ID of the API key
- `key_name` (String) Name of the API key
- `role` (String) Role of the API key. This role can be either `GRAPH_ADMIN`, `CONTRIBUTOR`, `DOCUMENTER`, `OBSERVER` or `CONSUMER`
- `token` (String) Authentication token of the API key. This value is only fully available when creating the API key, the current value is partially masked
