package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/client"
)

var _ datasource.DataSource = &OrganizationDataSource{}

type OrganizationDataSource struct {
	client *client.ApolloClient
}

type OrganizationDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	IsOnTrial        types.Bool   `tfsdk:"is_on_trial"`
	IsOnExpiredTrial types.Bool   `tfsdk:"is_on_expired_trial"`
	IsLocked         types.Bool   `tfsdk:"is_locked"`
}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

func (d *OrganizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Organization data source", // TODO: change this
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"is_on_trial": schema.BoolAttribute{
				Computed: true,
			},
			"is_on_expired_trial": schema.BoolAttribute{
				Computed: true,
			},
			"is_locked": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.client.GetOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get organization", // TODO: change this
			"Unable to get the current organization",
		)
		return
	}

	data.Id = types.StringValue(org.Id)
	data.Name = types.StringValue(org.Name)
	data.IsOnTrial = types.BoolValue(org.IsOnTrial)
	data.IsOnExpiredTrial = types.BoolValue(org.IsOnExpiredTrial)
	data.IsLocked = types.BoolValue(org.IsLocked)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
