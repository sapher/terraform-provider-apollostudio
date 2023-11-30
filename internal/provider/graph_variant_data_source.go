package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var _ datasource.DataSource = &GraphVariantDataSource{}

type GraphVariantDataSource struct {
	client *client.ApolloClient
}

type GraphVariantDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewGraphVariantDataSource() datasource.DataSource {
	return &GraphVariantDataSource{}
}

func (d *GraphVariantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_variant"
}

func (d *GraphVariantDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provide details about a specific graph variant",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the variant",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the variant",
				Computed:    true,
			},
		},
	}
}

func (d *GraphVariantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ApolloClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ApolloClient got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GraphVariantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GraphVariantDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	graphVariant, err := d.client.GetGraphVariant(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get graph variant",
			fmt.Sprintf("Failed to get graph variant: %s", err.Error()),
		)
		return
	}

	data.Id = types.StringValue(graphVariant.Id)
	data.Name = types.StringValue(graphVariant.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
