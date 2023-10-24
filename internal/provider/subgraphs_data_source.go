package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var _ datasource.DataSource = &SubGraphsDataSource{}

type SubGraphsDataSource struct {
	client *client.ApolloClient
}

type SubGraphDataSourceModel struct {
	Name                types.String       `tfsdk:"name"`
	Revision            types.String       `tfsdk:"revision"`
	Url                 types.String       `tfsdk:"url"`
	ActivePartialSchema PartialSchemaModel `tfsdk:"active_partial_schema"`
}

type SubGraphsDataSourceModel struct {
	GraphId        types.String              `tfsdk:"graph_id"`
	VariantName    types.String              `tfsdk:"variant_name"`
	IncludeDeleted types.Bool                `tfsdk:"include_deleted"`
	SubGraphs      []SubGraphDataSourceModel `tfsdk:"subgraphs"`
}

func NewSubGraphsDataSource() datasource.DataSource {
	return &SubGraphsDataSource{}
}

func (d *SubGraphsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraphs"
}

func (d *SubGraphsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provide details about a specific variant's subgraphs",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "ID of the graph",
				Required:    true,
			},
			"variant_name": schema.StringAttribute{
				Description: "Name of the variant", // TODO: could be put optional, if the default value is current
				Required:    true,
			},
			"include_deleted": schema.BoolAttribute{
				Description: "Boolean indicating if deleted subgraphs should be included in the result",
				Optional:    true,
			},
			"subgraphs": schema.ListNestedAttribute{
				Description: "List of graph subgraphs",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the subgraph",
							Computed:    true,
						},
						"revision": schema.StringAttribute{
							Description: "Revision of the subgraph",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "Routing URL of the subgraph",
							Computed:    true,
						},
						"active_partial_schema": schema.SingleNestedAttribute{
							Description: "Provide details about the subgraph active schema",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"sdl": schema.StringAttribute{
									Description: "SDL of the active schema",
									Computed:    true,
								},
								"created_at": schema.StringAttribute{
									Description: "Creation date of the active schema",
									Computed:    true,
								},
								"is_live": schema.BoolAttribute{
									Description: "Boolean indicating if the active schema is live",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *SubGraphsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubGraphsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubGraphsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subgraphs, err := d.client.GetSubGraphs(ctx, data.GraphId.ValueString(), data.VariantName.ValueString(), data.IncludeDeleted.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get subgraphs for the given graph",
			fmt.Sprintf("Failed to get subgraphs for the given graph: %s", err.Error()),
		)
		return
	}

	for _, subgraph := range subgraphs {
		data.SubGraphs = append(data.SubGraphs, SubGraphDataSourceModel{
			Name:     types.StringValue(subgraph.Name),
			Revision: types.StringValue(subgraph.Revision),
			Url:      types.StringValue(subgraph.Url),
			ActivePartialSchema: PartialSchemaModel{
				Sdl:       types.StringValue(subgraph.ActivePartialSchema.Sdl),
				CreatedAt: types.StringValue(subgraph.ActivePartialSchema.CreatedAt),
				IsLive:    types.BoolValue(subgraph.ActivePartialSchema.IsLive),
			},
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
