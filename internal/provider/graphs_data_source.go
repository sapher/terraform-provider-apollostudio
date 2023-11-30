package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
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
		Description: "Provide details about a specific organization's graphs",
		Attributes: map[string]schema.Attribute{
			"graphs": schema.ListNestedAttribute{
				Description: "List of graphs",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the graph",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the graph",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the graph",
							Computed:    true,
						},
						"graph_type": schema.StringAttribute{
							Description: "Type of the graph. This can be one of: `CLASSIC`, `CLOUD_SUPERGRAPH`",
							Computed:    true,
						},
						"reporting_enabled": schema.BoolAttribute{
							Description: "Boolean indicating if reporting is enabled for the graph",
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
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ApolloClient got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
			"Failed to get graphs",
			fmt.Sprintf("Failed to get graphs: %s", err.Error()),
		)
		return
	}

	for _, graph := range graphs {
		data.Graphs = append(data.Graphs, GraphDataSourceModel{
			Id:               types.StringValue(graph.Id),
			Name:             types.StringValue(graph.Name),
			Description:      types.StringValue(graph.Description),
			GraphType:        types.StringValue(graph.GraphType),
			ReportingEnabled: types.BoolValue(graph.ReportingEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
