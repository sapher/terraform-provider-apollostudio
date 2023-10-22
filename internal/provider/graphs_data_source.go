package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &GraphsDataSource{}

type GraphsDataSource struct {
	client *client.ApolloClient
}

type GraphsDataSourceModel struct {
	Graphs []GraphDataSourceModel `tfsdk:"graphs"`
}

func NewGraphsDataSource() datasource.DataSource {
	return &GraphsDataSource{}
}

func (d *GraphsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graphs"
}

func (d *GraphsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all graphs of the organization", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"graphs": schema.ListNestedAttribute{
				Description: "List of graphs",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Graph ID",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Graph name",
							Computed:    true,
						},
						"title": schema.StringAttribute{
							Description: "Graph title",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Graph description",
							Computed:    true,
						},
						"graph_type": schema.StringAttribute{
							Description: "Graph type",
							Computed:    true,
						},
						"reporting_enabled": schema.BoolAttribute{
							Description: "Whether reporting is enabled for the graph",
							Computed:    true,
						},
						"account_id": schema.StringAttribute{
							Description: "Account ID",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *GraphsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GraphsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GraphsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	graphs, err := d.client.GetGraphs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get graph", // TODO: change this
			"Unable to get graph", // TODO: change this
		)
		return
	}

	for _, graph := range graphs {
		data.Graphs = append(data.Graphs, GraphDataSourceModel{
			Id:               types.StringValue(graph.Id),
			Name:             types.StringValue(graph.Name),
			Title:            types.StringValue(graph.Title),
			Description:      types.StringValue(graph.Description),
			GraphType:        types.StringValue(graph.GraphType),
			ReportingEnabled: types.BoolValue(graph.ReportingEnabled),
			AccountId:        types.StringValue(graph.AccountId),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
