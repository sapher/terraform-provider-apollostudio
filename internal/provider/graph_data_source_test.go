package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
					data "apollostudio_graph" "this" {
						id = "testacc-terraform"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "id", "testacc-terraform"),
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "name", "testacc-terraform"),
					resource.TestCheckResourceAttr("data.apollostudio_graph.this", "description", "graph dedicated for terraform provider acceptance testing"),
				),
			},
		},
	})
}
