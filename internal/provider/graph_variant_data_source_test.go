package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphVariantDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
					data "apollostudio_graph_variant" "this" {
						id = "testacc-terraform@current"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_graph_variant.this", "name", "current"),
				),
			},
		},
	})
}
