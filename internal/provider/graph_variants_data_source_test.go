package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphVariantsDataSource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccSubgraphConfig(graphId, testAccSchema) + `
data "apollostudio_graph_variants" "this" {
  graph_id   = apollostudio_graph.fixture.id
  depends_on = [apollostudio_subgraph.fixture]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph_variants.this", "variants.#", "1"),
					resource.TestCheckResourceAttr("data.apollostudio_graph_variants.this", "variants.0.name", testAccVariantName),
				),
			},
		},
	})
}
