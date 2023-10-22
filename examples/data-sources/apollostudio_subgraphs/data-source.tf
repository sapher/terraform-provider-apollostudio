data "apollostudio_subgraphs" "this" {
  graph_id        = "your-graph-id"
  variant_name    = "your-variant-name"
  include_deleted = false
}
