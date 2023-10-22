package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type PartialSchemaModel struct {
	Sdl       types.String `tfsdk:"sdl"`
	CreatedAt types.String `tfsdk:"created_at"`
	IsLive    types.Bool   `tfsdk:"is_live"`
}

type IdendityModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
