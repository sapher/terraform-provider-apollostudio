package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphVariantsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
					data "apollostudio_graph_variants" "this" {
						graph_id = "testacc-terraform"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph_variants.this", "variants.#", "1"),
				),
			},
		},
	})
}
