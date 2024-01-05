package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubgraphsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
					data "apollostudio_subgraphs" "this" {
						graph_id        = "testacc-terraform"
						variant_name    = "current"
						include_deleted = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_subgraphs.this", "subgraphs.#", "1"),
					resource.TestCheckResourceAttr("data.apollostudio_subgraphs.this", "subgraphs.0.name", "countries"),
				),
			},
		},
	})
}
