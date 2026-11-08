package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphsDataSource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccGraphConfig(graphId) + `
data "apollostudio_graphs" "this" {
  depends_on = [apollostudio_graph.fixture]
}
`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.apollostudio_graphs.this", "graphs.*", map[string]string{
					"id":   graphId,
					"name": graphId,
				}),
			},
		},
	})
}
