resource "apollostudio_subgraph" "this" {
  graph_id      = "your-graph-id"
  variant_name  = "your-variant-name"
  subgraph_name = "your-subgraph-name"
  schema        = file("schema.graphql")
  url           = "https://your-api-url.com/graphql"
  revision      = "your-revision"
}
