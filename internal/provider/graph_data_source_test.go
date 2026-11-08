package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphDataSource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccGraphConfig(graphId) + `
data "apollostudio_graph" "this" {
  id = apollostudio_graph.fixture.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "id", graphId),
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "name", graphId),
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "description", "graph dedicated for terraform provider acceptance testing"),
					resource.TestCheckResourceAttrSet("data.apollostudio_graph.this", "graph_type"),
				),
			},
		},
	})
}
