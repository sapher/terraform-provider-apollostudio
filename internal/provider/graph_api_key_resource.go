package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var (
	_ resource.Resource              = &GraphApiKeyResource{}
	_ resource.ResourceWithConfigure = &GraphApiKeyResource{}
)

type GraphApiKeyResource struct {
	client *client.ApolloClient
}

type GraphApiKeyResourceModel struct {
	GraphId   types.String  `tfsdk:"graph_id"`
	Id        types.String  `tfsdk:"id"`
	KeyName   types.String  `tfsdk:"key_name"`
	Role      types.String  `tfsdk:"role"`
	Token     types.String  `tfsdk:"token"`
	CreatedAt types.String  `tfsdk:"created_at"`
	CreatedBy IdendityModel `tfsdk:"created_by"`
}

func NewGraphApiKeyResource() resource.Resource {
	return &GraphApiKeyResource{}
}

func (r *GraphApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graph_api_key"
}

func (r *GraphApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage an API key for a specific graph",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "ID of the graph",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "ID of the API key",
				Computed:    true,
			},
			"key_name": schema.StringAttribute{
				Description: "Name of the API key",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "Role of the API key. This role can be either `GRAPH_ADMIN`, `CONTRIBUTOR`, `DOCUMENTER`, `OBSERVER` or `CONSUMER`",
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "Authentication token of the API key. This value is only fully available when creating the API key, the current value is partially masked",
				Computed:    true,
				Sensitive:   true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation date of the API key",
				Computed:    true,
			},
			"created_by": schema.SingleNestedAttribute{
				Description: "Creator of the API key",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID of the entity who created the key",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "Name of the entity who created the key",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *GraphApiKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *GraphApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state GraphApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed graph api key from Apollo Studio
	apiKey, err := r.client.GetGraphApiKey(ctx, state.GraphId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get graph api key",
			fmt.Sprintf("Failed to get graph api key: %s", err.Error()),
		)
		return
	}

	// If not found
	if apiKey.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Override state with refreshed values
	state.Id = types.StringValue(apiKey.Id)
	state.Role = types.StringValue(apiKey.Role)
	state.KeyName = types.StringValue(apiKey.KeyName)
	state.CreatedAt = types.StringValue(apiKey.CreatedAt)
	state.CreatedBy = IdendityModel{
		Id:   types.StringValue(apiKey.CreatedBy.Id),
		Name: types.StringValue(apiKey.CreatedBy.Name),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Return values from plan
	var plan GraphApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the API key
	apiKey, err := r.client.CreateGraphApiKey(ctx, plan.GraphId.ValueString(), plan.KeyName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating graph api key",
			"Could not create graph api key, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate response
	plan.Id = types.StringValue(apiKey.Id)
	plan.Role = types.StringValue(apiKey.Role)
	plan.Token = types.StringValue(apiKey.Token)
	plan.CreatedAt = types.StringValue(apiKey.CreatedAt)
	plan.CreatedBy = IdendityModel{
		Id:   types.StringValue(apiKey.CreatedBy.Id),
		Name: types.StringValue(apiKey.CreatedBy.Name),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Return values from plan
	var plan GraphApiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state GraphApiKeyResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the API key
	err := r.client.RenameGraphApiKey(ctx, state.GraphId.ValueString(), state.Id.ValueString(), plan.KeyName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating graph api key",
			"Could not update graph api key "+plan.Id.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate response
	plan.Id = state.Id
	plan.Role = state.Role
	plan.Token = state.Token
	plan.CreatedAt = state.CreatedAt
	plan.CreatedBy = state.CreatedBy

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *GraphApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state GraphApiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the API key
	err := r.client.RemoveGraphApiKey(ctx, state.GraphId.ValueString(), state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete graph api key",
			"Failed to delete graph api key "+state.Id.ValueString()+", unexpected error: "+err.Error(),
		)
		return
	}
}
