package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgDataSource(t *testing.T) {
	// Retrieve the current organization from the client
	cl := testClient()
	org, _ := cl.GetOrganization(context.Background())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "apollostudio_org" "current" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_org.current", "id", org.Id),
					resource.TestCheckResourceAttr("data.apollostudio_org.current", "name", org.Name),
				),
			},
		},
	})
}
