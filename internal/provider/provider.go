package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ provider.Provider = &ApolloProvider{}

type ApolloProvider struct {
	version string
}

type ApolloProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
	OrgId  types.String `tfsdk:"org_id"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApolloProvider{
			version: version,
		}
	}
}

func (p *ApolloProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apollostudio"
	resp.Version = p.version
}

func (p *ApolloProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Apollo Studio",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Host of the Apollo GraphQL API. Defaults to https://graphql.api.apollographql.com/api/graphql",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for the Apollo GraphQL API. Can also be set via the APOLLO_KEY environment variable",
				Optional:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID on Apollo Studio. Can also be set via the APOLLO_ORG_ID environment variable",
				Required:    true,
			},
		},
	}
}

func (p *ApolloProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ApolloProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	host := "https://graphql.api.apollographql.com/api/graphql"
	apiKey := os.Getenv("APOLLO_KEY")
	orgId := os.Getenv("APOLLO_ORG_ID")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if !config.OrgId.IsNull() {
		orgId = config.OrgId.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Apollo host",
			"Please set the host to the Apollo host",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Apollo API key",
			"Please set the api_key so the client can authenticate to the Apollo API",
		)
	}

	if orgId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("org_id"),
			"Missing Apollo org ID",
			"Please set the org_id to the Apollo org ID",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "apollo_host", host)
	ctx = tflog.SetField(ctx, "apollo_api_key", apiKey)
	ctx = tflog.SetField(ctx, "apollo_org_id", orgId)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "apollo_api_key")

	tflog.Info(ctx, "Configured Apollo provider")

	client := client.NewClient(host, apiKey, orgId)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ApolloProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGraphApiKeyResource,
	}
}

func (p *ApolloProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMeDataSource,
		NewOrganizationDataSource,
		NewGraphDataSource,
		NewGraphsDataSource,
		NewGraphVariantDataSource,
		NewGraphVariantsDataSource,
		NewGraphApiKeysDataSource,
		NewSubGraphsDataSource,
	}
}
