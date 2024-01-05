package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGraphResource(t *testing.T) {
	randomId := uuid.New().String()[0:8]
	graphId := fmt.Sprintf("test-%s", randomId)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `resource "apollostudio_graph" "this" {
					id = "` + graphId + `"
					name = "` + graphId + `"
					description = "Test Graph"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("apollostudio_graph.this", "id", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "name", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "description", "Test Graph"),
				),
			},
			// ImportState
			{
				ResourceName:      "apollostudio_graph.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and read testing
			{
				Config: providerConfig + `resource "apollostudio_graph" "this" {
					id = "` + graphId + `"
					name = "` + graphId + `"
					description = "Test Graph Updated"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("apollostudio_graph.this", "id", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "name", graphId),
					resource.TestCheckResourceAttr("apollostudio_graph.this", "description", "Test Graph Updated"),
				),
			},
			// Delete is automatically done by the framework
		},
	})
}
