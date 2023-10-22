package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &GraphApiKeysDataSource{}

type GraphApiKeysDataSource struct {
	client *client.ApolloClient
}

type GraphApiKeyDataSourceModel struct {
	Id        types.String  `tfsdk:"id"`
	KeyName   types.String  `tfsdk:"key_name"`
	Role      types.String  `tfsdk:"role"`
	Token     types.String  `tfsdk:"token"`
	CreatedAt types.String  `tfsdk:"created_at"`
	CreatedBy IdendityModel `tfsdk:"created_by"`
}

type GraphApiKeysDataSourceModel struct {
	GraphId      types.String                 `tfsdk:"graph_id"`
	GraphApiKeys []GraphApiKeyDataSourceModel `tfsdk:"api_keys"`
}

func NewGraphApiKeysDataSource() datasource.DataSource {
	return &GraphApiKeysDataSource{}
}

func (d *GraphApiKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_api_keys"
}

func (d *GraphApiKeysDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of api keys for a given graph", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "Graph ID",
				Required:    true,
			},
			"api_keys": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Api key ID",
							Computed:    true,
						},
						"key_name": schema.StringAttribute{
							Description: "Key name",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "Key role",
							Computed:    true,
						},
						"token": schema.StringAttribute{
							Description: "Key token",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Key creation date",
							Computed:    true,
						},
						"created_by": schema.SingleNestedAttribute{
							Description: "Creator of the key",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Identity ID",
									Computed:    true,
								},
								"name": schema.StringAttribute{
									Description: "Identity name",
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

func (d *GraphApiKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GraphApiKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GraphApiKeysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	graphApiKeys, err := d.client.GetGraphApiKeys(ctx, data.GraphId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get graph", // TODO: change this
			"Unable to get graph", // TODO: change this
		)
		return
	}

	for _, graphApiKey := range graphApiKeys {
		data.GraphApiKeys = append(data.GraphApiKeys, GraphApiKeyDataSourceModel{
			Id:        types.StringValue(graphApiKey.Id),
			KeyName:   types.StringValue(graphApiKey.KeyName),
			Role:      types.StringValue(graphApiKey.Role),
			Token:     types.StringValue(graphApiKey.Token),
			CreatedAt: types.StringValue(graphApiKey.CreatedAt),
			CreatedBy: IdendityModel{
				Id:   types.StringValue(graphApiKey.CreatedBy.Id),
				Name: types.StringValue(graphApiKey.CreatedBy.Name),
			},
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
