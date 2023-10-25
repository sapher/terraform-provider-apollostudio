package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IdendityModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
