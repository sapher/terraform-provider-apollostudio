package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccGraphApiKeyConfig(graphId, keyName string) string {
	return providerConfig + testAccGraphConfig(graphId) + fmt.Sprintf(`
resource "apollostudio_graph_api_key" "this" {
  graph_id = apollostudio_graph.fixture.id
  key_name = %q
}

data "apollostudio_graph_api_keys" "this" {
  graph_id   = apollostudio_graph.fixture.id
  depends_on = [apollostudio_graph_api_key.this]
}
`, keyName)
}

func TestAccGraphApiKeyResource(t *testing.T) {
	graphId := testAccGraphId()
	const address = "apollostudio_graph_api_key.this"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create, and list it through the data source
			{
				Config: testAccGraphApiKeyConfig(graphId, "tfacc-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "key_name", "tfacc-key"),
					resource.TestCheckResourceAttrSet(address, "id"),
					resource.TestCheckResourceAttrSet(address, "token"),
					resource.TestCheckResourceAttrSet(address, "role"),
					resource.TestCheckTypeSetElemNestedAttrs("data.apollostudio_graph_api_keys.this", "api_keys.*", map[string]string{
						"key_name": "tfacc-key",
					}),
				),
			},
			// Rename in place
			{
				Config: testAccGraphApiKeyConfig(graphId, "tfacc-key-renamed"),
				Check:  resource.TestCheckResourceAttr(address, "key_name", "tfacc-key-renamed"),
			},
		},
	})
}
