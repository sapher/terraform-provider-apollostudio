package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
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
		Description: "Provide details about a specific graph",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the graph",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the graph",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Title of the graph",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the graph",
				Computed:    true,
			},
			"graph_type": schema.StringAttribute{
				Description: "Type of the graph",
				Computed:    true,
			},
			"reporting_enabled": schema.BoolAttribute{
				Description: "Boolean indicating if reporting is enabled for the graph",
				Computed:    true,
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
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ApolloClientn got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
			"Failed to get graph",
			fmt.Sprintf("Failed to get graph: %s", err.Error()),
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
