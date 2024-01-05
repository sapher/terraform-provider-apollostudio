package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

var (
	_ resource.Resource                = &SubGraphResource{}
	_ resource.ResourceWithConfigure   = &SubGraphResource{}
	_ resource.ResourceWithImportState = &SubGraphResource{}
)

type SubGraphResource struct {
	client *client.ApolloClient
}

type SubGraphResourceModel struct {
	GraphId     types.String `tfsdk:"graph_id"`
	VariantName types.String `tfsdk:"variant_name"`
	Name        types.String `tfsdk:"name"`
	Schema      types.String `tfsdk:"schema"`
	Url         types.String `tfsdk:"url"`
	Revision    types.String `tfsdk:"revision"`
}

func NewSubGraphResource() resource.Resource {
	return &SubGraphResource{}
}

func (r *SubGraphResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subgraph"
}

func (r *SubGraphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a subgraph",
		Attributes: map[string]schema.Attribute{
			"graph_id": schema.StringAttribute{
				Description: "ID of the graph",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 40),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]+$`),
						"must starts with a letter and contains only letters, numbers, and dashes",
					),
				},
			},
			"variant_name": schema.StringAttribute{
				Description: "Name of the subgraph variant",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the subgraph",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema": schema.StringAttribute{
				Description: "Schema of the subgraph variant",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL of the subgraph variant",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision": schema.StringAttribute{
				Description: "Revision of the subgraph variant",
				Computed:    true,
				// Revision must update when schema change
			},
		},
	}
}

func (r *SubGraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *SubGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Return values from plan
	var plan SubGraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Publish the subgraph
	err := r.client.PublishSubGraph(
		ctx,
		plan.GraphId.ValueString(),
		plan.VariantName.ValueString(),
		plan.Name.ValueString(),
		plan.Schema.ValueString(),
		plan.Url.ValueString(),
		"1", // plan.Revision.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create subgraph",
			fmt.Sprintf("Failed to create subgraph: %s", err.Error()),
		)
		return
	}

	// Set the revision to 1
	plan.Revision = types.StringValue("1")
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Return values from state
	var state SubGraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the subgraph
	subgraph, err := r.client.GetSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get subgraph",
			fmt.Sprintf("Failed to get subgraph: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate response
	state.Url = types.StringValue(subgraph.Url)
	state.Revision = types.StringValue(subgraph.Revision)
	state.Schema = types.StringValue(subgraph.ActivePartialSchema.Sdl)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SubGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Return values from plan
	var plan SubGraphResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Return values from state
	var state SubGraphResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate Schema
	workflowId, err := r.client.SubmitSubgraphCheck(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString(), plan.Schema.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to validate subgraph schema",
			fmt.Sprintf("Failed to validate subgraph schema: %s", err.Error()),
		)
		return
	}

	tflog.Warn(ctx, fmt.Sprintf("Workflow ID: %s", workflowId))

	validationResults, err := r.client.CheckWorkflow(ctx, state.GraphId.ValueString(), workflowId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to validate subgraph schema",
			fmt.Sprintf("Failed to validate subgraph schema: %s", err.Error()),
		)
		return
	}

	// Check if validation results contains errors
	if len(validationResults) > 0 {
		for _, result := range validationResults {
			for _, message := range result.Messages {
				resp.Diagnostics.AddError(
					"Failed to validate subgraph schema",
					fmt.Sprintf("Failed to validate subgraph schema: %s", message),
				)
			}
		}
		return
	}

	// Update schema
	if plan.Schema.ValueString() != state.Schema.ValueString() {
		err := r.client.PublishSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString(), plan.Schema.ValueString(), state.Url.ValueString(), state.Revision.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update subgraph schema",
				fmt.Sprintf("Failed to update subgraph schema: %s", err.Error()),
			)
			return
		}
	}

	// Get the subgraph
	subgraph, err := r.client.GetSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get subgraph",
			fmt.Sprintf("Failed to get subgraph: %s", err.Error()),
		)
		return
	}

	// Map response body to schema and populate response
	plan.Revision = types.StringValue(subgraph.Revision)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SubGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Return values from state
	var state SubGraphResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the graph
	err := r.client.RemoveSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete subgraph",
			fmt.Sprintf("Failed to delete subgraph: %s", err.Error()),
		)
		return
	}
}

func (r *SubGraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, fmt.Sprintf("Import subgraph: %s", req.ID))
	pattern := "^([a-zA-Z0-9_-]+)@([a-zA-Z0-9_-]+):([a-zA-Z0-9_-]+)$"
	matched, _ := regexp.MatchString(pattern, req.ID)
	if !matched {
		resp.Diagnostics.AddError(
			"Invalid subgraph ID",
			fmt.Sprintf("Invalid subgraph ID: %s", req.ID),
		)
		return
	}

	// extract each part of the string
	re := regexp.MustCompile(pattern)
	matchs := re.FindStringSubmatch(req.ID)
	graphId := matchs[1]
	variantName := matchs[2]
	name := matchs[3]

	// Get the subgraph
	subgraph, err := r.client.GetSubGraph(ctx, graphId, variantName, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get subgraph",
			fmt.Sprintf("Failed to get subgraph: %s", err.Error()),
		)
		return
	}

	// Return values from plan
	var plan SubGraphResourceModel
	plan.GraphId = types.StringValue(graphId)
	plan.VariantName = types.StringValue(variantName)
	plan.Name = types.StringValue(name)
	plan.Schema = types.StringValue(subgraph.ActivePartialSchema.Sdl)
	plan.Url = types.StringValue(subgraph.Url)
	plan.Revision = types.StringValue(subgraph.Revision)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
