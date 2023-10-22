package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &MeDataSource{}

type MeDataSource struct {
	client *client.ApolloClient
}

type MeDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewMeDataSource() datasource.DataSource {
	return &MeDataSource{}
}

func (d *MeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_me"
}

func (d *MeDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Current authenticated user", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "User name",
				Computed:    true,
			},
		},
	}
}

func (d *MeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ApolloClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ApolloClientn got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *MeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.client.GetMe(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get the current authenticated user", // TODO: change this
			"Unable to get date for the current user",
		)
		return
	}

	data.Id = types.StringValue(me.Id)
	data.Name = types.StringValue(me.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
