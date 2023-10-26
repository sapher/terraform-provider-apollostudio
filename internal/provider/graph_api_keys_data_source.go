package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var _ datasource.DataSource = &GraphApiKeysDataSource{}

type GraphApiKeysDataSource struct {
	client *client.ApolloClient
}

type GraphApiKeyDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	KeyName   types.String `tfsdk:"key_name"`
	Role      types.String `tfsdk:"role"`
	Token     types.String `tfsdk:"token"`
	CreatedAt types.String `tfsdk:"created_at"`
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
		Description: "Provide details about a specific graph's API keys. Beware that the API key token is partially masked when read, it's only available at creation time.",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "ID of the graph linked to the API keys",
				Required:    true,
			},
			"api_keys": schema.ListNestedAttribute{
				Description: "List of API keys",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the API key",
							Computed:    true,
						},
						"key_name": schema.StringAttribute{
							Description: "Name of the API key",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "Role of the API key. This role can be either `GRAPH_ADMIN`, `CONTRIBUTOR`, `DOCUMENTER`, `OBSERVER` or `CONSUMER`",
							Computed:    true,
						},
						"token": schema.StringAttribute{
							Description: "Authentication token of the API key. This value is only fully available when creating the API key, the current value is partially masked",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Creation date of the API key",
							Computed:    true,
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
			"Failed to get graph",
			fmt.Sprintf("Failed to get graph: %s", err.Error()),
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
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
