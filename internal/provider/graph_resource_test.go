package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccGraphResourceConfig(graphId, name, description string) string {
	return providerConfig + fmt.Sprintf(`
resource "apollostudio_graph" "this" {
  id          = %q
  name        = %q
  description = %q
}
`, graphId, name, description)
}

func TestAccGraphResource(t *testing.T) {
	graphId := testAccGraphId()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccGraphResourceConfig(graphId, graphId, "Test Graph"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollostudio_graph.this", "id", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "name", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "description", "Test Graph"),
				),
			},
			// Import
			{
				ResourceName:      "apollostudio_graph.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update name and description in place
			{
				Config: testAccGraphResourceConfig(graphId, graphId+"-renamed", "Test Graph Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollostudio_graph.this", "id", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "name", graphId+"-renamed"),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "description", "Test Graph Updated"),
				),
			},
			// Delete is automatically done by the framework
		},
	})
}
