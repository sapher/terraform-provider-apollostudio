package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &GraphDataSource{}

type GraphDataSource struct {
	client *client.ApolloClient
}

type GraphDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	GraphType        types.String `tfsdk:"graph_type"`
	ReportingEnabled types.Bool   `tfsdk:"reporting_enabled"`
}

func NewGraphDataSource() datasource.DataSource {
	return &GraphDataSource{}
}

func (d *GraphDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph"
}

func (d *GraphDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Graph data source", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"title": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"graph_type": schema.StringAttribute{
				Computed: true,
			},
			"reporting_enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *GraphDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GraphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GraphDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	graph, err := d.client.GetGraph(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get graph", // TODO: change this
			"Unable to get graph", // TODO: change this
		)
		return
	}

	data.Name = types.StringValue(graph.Name)
	data.Title = types.StringValue(graph.Title)
	data.Description = types.StringValue(graph.Description)
	data.GraphType = types.StringValue(graph.GraphType)
	data.ReportingEnabled = types.BoolValue(graph.ReportingEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
