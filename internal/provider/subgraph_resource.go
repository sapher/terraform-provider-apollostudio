package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

// subgraphCheckTimeout bounds how long we wait for an Apollo schema check to complete.
const subgraphCheckTimeout = 10 * time.Minute

// subgraphImportIdRegexp matches import IDs of the form <graph-id>@<variant-name>:<subgraph-name>.
var subgraphImportIdRegexp = regexp.MustCompile(`^([a-zA-Z0-9_-]+)@([a-zA-Z0-9_-]+):([a-zA-Z0-9_-]+)$`)

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
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ApolloClient got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SubGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubGraphResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	graphId := plan.GraphId.ValueString()
	variantName := plan.VariantName.ValueString()

	// Checks can only run against an existing variant; the first publish creates the variant.
	_, err := r.client.GetGraphVariant(ctx, graphId+"@"+variantName)
	switch {
	case client.IsNotFound(err):
		tflog.Info(ctx, fmt.Sprintf("Variant %s@%s does not exist yet, skipping schema check", graphId, variantName))
	case err != nil:
		resp.Diagnostics.AddError("Failed to get graph variant", fmt.Sprintf("Failed to get graph variant %s@%s: %s", graphId, variantName, err.Error()))
		return
	default:
		r.checkSchema(ctx, plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Publish the subgraph
	compositionErrors, err := r.client.PublishSubGraph(ctx, graphId, variantName, plan.Name.ValueString(), plan.Schema.ValueString(), plan.Url.ValueString(), "1")
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subgraph", fmt.Sprintf("Failed to create subgraph: %s", err.Error()))
		return
	}

	// The subgraph exists from here on: always record it in state, even if composition failed.
	plan.Revision = types.StringValue("1")
	r.refreshRevision(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	addCompositionErrors(compositionErrors, &resp.Diagnostics)
}

func (r *SubGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubGraphResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subgraph, err := r.client.GetSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString())
	if client.IsNotFound(err) {
		tflog.Warn(ctx, fmt.Sprintf("Subgraph %s@%s:%s not found, removing it from state", state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to get subgraph", fmt.Sprintf("Failed to get subgraph: %s", err.Error()))
		return
	}

	state.Url = types.StringValue(subgraph.Url)
	state.Revision = types.StringValue(subgraph.Revision)
	state.Schema = types.StringValue(subgraph.ActivePartialSchema.Sdl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SubGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubGraphResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SubGraphResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaChanged := !plan.Schema.Equal(state.Schema)
	urlChanged := !plan.Url.Equal(state.Url)

	if schemaChanged {
		r.checkSchema(ctx, plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plan.Revision = state.Revision
	var compositionErrors []string
	if schemaChanged || urlChanged {
		var err error
		compositionErrors, err = r.client.PublishSubGraph(ctx, plan.GraphId.ValueString(), plan.VariantName.ValueString(), plan.Name.ValueString(), plan.Schema.ValueString(), plan.Url.ValueString(), state.Revision.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to update subgraph", fmt.Sprintf("Failed to update subgraph: %s", err.Error()))
			return
		}
	}

	r.refreshRevision(ctx, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	addCompositionErrors(compositionErrors, &resp.Diagnostics)
}

func (r *SubGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubGraphResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveSubGraph(ctx, state.GraphId.ValueString(), state.VariantName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete subgraph", fmt.Sprintf("Failed to delete subgraph: %s", err.Error()))
		return
	}
}

// ImportState accepts IDs of the form <graph-id>@<variant-name>:<subgraph-name>; Read fills in the rest.
func (r *SubGraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	matches := subgraphImportIdRegexp.FindStringSubmatch(req.ID)
	if matches == nil {
		resp.Diagnostics.AddError(
			"Invalid subgraph ID",
			fmt.Sprintf("Invalid subgraph ID %q, expected <graph-id>@<variant-name>:<subgraph-name>", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("graph_id"), matches[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variant_name"), matches[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), matches[3])...)
}

// checkSchema runs an Apollo subgraph check for the planned schema and reports its outcome as diagnostics.
func (r *SubGraphResource) checkSchema(ctx context.Context, plan SubGraphResourceModel, diags *diag.Diagnostics) {
	graphId := plan.GraphId.ValueString()

	workflowId, err := r.client.SubmitSubgraphCheck(ctx, graphId, plan.VariantName.ValueString(), plan.Name.ValueString(), plan.Schema.ValueString())
	if err != nil {
		diags.AddError("Failed to submit a graph validation check", fmt.Sprintf("Failed to submit a graph validation check: %s", err.Error()))
		return
	}

	checkCtx, cancel := context.WithTimeout(ctx, subgraphCheckTimeout)
	defer cancel()

	result, err := r.client.CheckWorkflow(checkCtx, graphId, workflowId)
	if err != nil {
		diags.AddError("Failed to check the workflow of a graph validation check", fmt.Sprintf("Failed to check the workflow of a graph validation check: %s", err.Error()))
		return
	}

	var errorsBuilder, warningsBuilder strings.Builder
	for _, task := range result.Tasks {
		for _, detail := range task.Details {
			message := fmt.Sprintf("%s : %s\n", task.TaskName, detail.Message)
			if detail.Level == client.LogLevelError {
				errorsBuilder.WriteString(message)
			} else {
				warningsBuilder.WriteString(message)
			}
		}
	}

	if warnings := strings.TrimSpace(warningsBuilder.String()); warnings != "" {
		diags.AddWarning("Warnings while validating subgraph schema", fmt.Sprintf("Warnings while validating subgraph schema:\n\n%s", warnings))
	}

	errs := strings.TrimSpace(errorsBuilder.String())
	if errs == "" && result.Failed() {
		errs = fmt.Sprintf("check workflow %s ended with status %s", workflowId, result.Status)
	}
	if errs != "" {
		diags.AddError("Failed to validate subgraph schema", fmt.Sprintf("Failed to validate subgraph schema:\n\n%s", errs))
	}
}

// refreshRevision reads the subgraph back to record the revision Apollo stored.
func (r *SubGraphResource) refreshRevision(ctx context.Context, model *SubGraphResourceModel, diags *diag.Diagnostics) {
	subgraph, err := r.client.GetSubGraph(ctx, model.GraphId.ValueString(), model.VariantName.ValueString(), model.Name.ValueString())
	if err != nil {
		diags.AddWarning("Failed to read subgraph revision", fmt.Sprintf("Subgraph was published but its revision could not be read back: %s", err.Error()))
		return
	}
	model.Revision = types.StringValue(subgraph.Revision)
}

func addCompositionErrors(compositionErrors []string, diags *diag.Diagnostics) {
	if len(compositionErrors) == 0 {
		return
	}
	diags.AddError(
		"Subgraph published but supergraph composition failed",
		fmt.Sprintf("The subgraph was stored by Apollo but the supergraph was not updated:\n\n%s", strings.Join(compositionErrors, "\n")),
	)
}
