package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &GraphVariantsDataSource{}

type GraphVariantsDataSource struct {
	client *client.ApolloClient
}

type GraphVariantsDataSourceModel struct {
	GraphId       types.String                  `tfsdk:"graph_id"`
	GraphVariants []GraphVariantDataSourceModel `tfsdk:"variants"`
}

func NewGraphVariantsDataSource() datasource.DataSource {
	return &GraphVariantsDataSource{}
}

func (d *GraphVariantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_variants"
}

func (d *GraphVariantsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Graph Variants data source", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Required: true,
			},
			"variants": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"has_supergraph_schema": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *GraphVariantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ApolloClient)

	if !ok {
		resp.Diagnostics.AddError(
			"invalid provider data",                          // TODO: change this
			"the provider data was not of the expected type", // TODO: change this
		)
		return
	}

	d.client = client
}

func (d *GraphVariantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GraphVariantsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	graphVariants, err := d.client.GetGraphVariants(ctx, data.GraphId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get graph", // TODO: change this
			"Unable to get graph", // TODO: change this
		)
		return
	}

	for _, graphVariant := range graphVariants {
		data.GraphVariants = append(data.GraphVariants, GraphVariantDataSourceModel{
			Id:                  types.StringValue(graphVariant.Id),
			Name:                types.StringValue(graphVariant.Name),
			HasSupergraphSchema: types.BoolValue(graphVariant.HasSupergraphSchema),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
