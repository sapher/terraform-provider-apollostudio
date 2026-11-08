package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSubgraphResource(t *testing.T) {
	graphId := testAccGraphId()
	const address = "apollostudio_subgraph.fixture"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create (first publish creates the variant, so no schema check runs)
			{
				Config: providerConfig + testAccSubgraphConfig(graphId, testAccSchema),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "graph_id", graphId),
					resource.TestCheckResourceAttr(address, "variant_name", testAccVariantName),
					resource.TestCheckResourceAttr(address, "name", testAccSubgraphName),
					resource.TestCheckResourceAttr(address, "url", testAccSubgraphUrl),
					resource.TestCheckResourceAttr(address, "schema", testAccSchema),
					resource.TestCheckResourceAttrSet(address, "revision"),
				),
			},
			// Import using <graph-id>@<variant-name>:<subgraph-name>
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[address]
					if !ok {
						return "", fmt.Errorf("%s not found in state", address)
					}
					return fmt.Sprintf("%s@%s:%s", rs.Primary.Attributes["graph_id"], rs.Primary.Attributes["variant_name"], rs.Primary.Attributes["name"]), nil
				},
			},
			// Update the schema: runs an Apollo check, then publishes
			{
				Config: providerConfig + testAccSubgraphConfig(graphId, testAccSchemaV2),
				Check:  resource.TestCheckResourceAttr(address, "schema", testAccSchemaV2),
			},
		},
	})
}
