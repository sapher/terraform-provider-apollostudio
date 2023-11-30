resource "apollostudio_subgraph" "this" {
  graph_id     = "your-graph-id"
  variant_name = "your-variant-name"
  name         = "your-subgraph-name"
  url          = "your-graphql-endpoint-url"
  schema       = file("path-to-your-schema.graphql")
}
