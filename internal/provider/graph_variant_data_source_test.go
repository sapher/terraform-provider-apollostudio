package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphVariantDataSource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Publishing a subgraph is what creates the variant.
				Config: providerConfig + testAccSubgraphConfig(graphId, testAccSchema) + `
data "apollostudio_graph_variant" "this" {
  id         = "${apollostudio_graph.fixture.id}@current"
  depends_on = [apollostudio_subgraph.fixture]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph_variant.this", "id", graphId+"@"+testAccVariantName),
					resource.TestCheckResourceAttr("data.apollostudio_graph_variant.this", "name", testAccVariantName),
				),
			},
		},
	})
}
