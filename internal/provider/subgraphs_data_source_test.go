package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubgraphsDataSource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccSubgraphConfig(graphId, testAccSchema) + `
data "apollostudio_subgraphs" "this" {
  graph_id        = apollostudio_graph.fixture.id
  variant_name    = "current"
  include_deleted = false
  depends_on      = [apollostudio_subgraph.fixture]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_subgraphs.this", "subgraphs.#", "1"),
					resource.TestCheckResourceAttr("data.apollostudio_subgraphs.this", "subgraphs.0.name", testAccSubgraphName),
					resource.TestCheckResourceAttr("data.apollostudio_subgraphs.this", "subgraphs.0.url", testAccSubgraphUrl),
				),
			},
		},
	})
}
