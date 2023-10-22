package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &GraphVariantDataSource{}

type GraphVariantDataSource struct {
	client *client.ApolloClient
}

type GraphVariantDataSourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	HasSupergraphSchema types.Bool   `tfsdk:"has_supergraph_schema"`
}

func NewGraphVariantDataSource() datasource.DataSource {
	return &GraphVariantDataSource{}
}

func (d *GraphVariantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_variant"
}

func (d *GraphVariantDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Specific Graph variant", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Graph variant ID",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Graph variant name",
				Computed:    true,
			},
			"has_supergraph_schema": schema.BoolAttribute{
				Description: "Whether the variant has a supergraph schema",
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
			fmt.Sprintf("Expected *client.ApolloClientn got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
			"failed to get graph", // TODO: change this
			"Unable to get graph", // TODO: change this
		)
		return
	}

	data.Id = types.StringValue(graphVariant.Id)
	data.Name = types.StringValue(graphVariant.Name)
	data.HasSupergraphSchema = types.BoolValue(graphVariant.HasSupergraphSchema)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
