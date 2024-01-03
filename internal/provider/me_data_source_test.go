package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMeDataSource(t *testing.T) {
	cl := testClient()
	me, _ := cl.GetMe(context.Background())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "apollostudio_me" "current" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.apollostudio_me.current", "id", me.Id),
					resource.TestCheckResourceAttr("data.apollostudio_me.current", "name", me.Name),
				),
			},
		},
	})
}
