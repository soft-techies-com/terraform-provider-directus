package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/terraform-provider-directus/internal/client"
)

// PolicyResource implements the directus_policy resource
type PolicyResource struct{ client *client.Directus }

// PolicyModel represents the policy resource model
type PolicyModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Icon        types.String   `tfsdk:"icon"`
	Description types.String   `tfsdk:"description"`
	IPAccess    []types.String `tfsdk:"ip_access"`
	EnforceTFA  types.Bool     `tfsdk:"enforce_tfa"`
	AdminAccess types.Bool     `tfsdk:"admin_access"`
	AppAccess   types.Bool     `tfsdk:"app_access"`
	// Permissions []types.Int64  `tfsdk:"permissions"`
	// Users       []types.String `tfsdk:"users"`
	// Roles []types.String `tfsdk:"roles"`
}

// NewPolicyResource returns a new policy resource
func NewPolicyResource() resource.Resource { return &PolicyResource{} }

// Metadata returns the resource type name
func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// Schema defines the schema for the resource
func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":          rschema.StringAttribute{Computed: true},
			"name":        rschema.StringAttribute{Required: true},
			"icon":        rschema.StringAttribute{Optional: true},
			"description": rschema.StringAttribute{Optional: true},
			"ip_access": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"enforce_tfa":  rschema.BoolAttribute{Optional: true},
			"admin_access": rschema.BoolAttribute{Optional: true},
			"app_access":   rschema.BoolAttribute{Optional: true},
		},
	}
}

// Configure configures the resource
func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Directus)
}
func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First payload: everything except roles
	payload := map[string]any{
		"name":         plan.Name.ValueString(),
		"icon":         plan.Icon.ValueString(),
		"description":  plan.Description.ValueString(),
		"ip_access":    expandStringList(plan.IPAccess),
		"enforce_tfa":  plan.EnforceTFA.ValueBool(),
		"admin_access": plan.AdminAccess.ValueBool(),
		"app_access":   plan.AppAccess.ValueBool(),
		// "roles":        expandStringList(plan.Roles),
		// roles intentionally skipped here
	}
	// Step 1: Create the policy
	httpResp, err := r.client.Request(ctx, http.MethodPost, "/policies", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error (create)", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error (parse create)", err.Error())
		return
	}

	r.readIntoState(ctx, &plan, resp, apiResp.Data)
}

// Read a policy
func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := "/policies/" + state.ID.ValueString()
	httpResp, err := r.client.Request(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	r.readIntoState(ctx, &state, resp, apiResp.Data)
}

// Update a policy
func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"name":         plan.Name.ValueString(),
		"icon":         plan.Icon.ValueString(),
		"description":  plan.Description.ValueString(),
		"ip_access":    expandStringList(plan.IPAccess),
		"enforce_tfa":  plan.EnforceTFA.ValueBool(),
		"admin_access": plan.AdminAccess.ValueBool(),
		"app_access":   plan.AppAccess.ValueBool(),
		// "permissions":  expandInt64List(plan.Permissions),
		// "users":        expandStringList(plan.Users),
		// "roles": expandStringList(plan.Roles),
	}

	var state PolicyModel
	req.State.Get(ctx, &state)
	url := "/policies/" + state.ID.ValueString()
	httpResp, err := r.client.Request(ctx, http.MethodPatch, url, payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	r.readIntoState(ctx, &plan, resp, apiResp.Data)
}

// Delete a policy
func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Request(ctx, http.MethodDelete, "/policies/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// helper: map API data into state
func (r *PolicyResource) readIntoState(ctx context.Context, plan *PolicyModel, resp interface{}, data map[string]any) {
	plan.ID = types.StringValue(str(data["id"]))
	plan.Name = strPtrToType(data["name"])
	plan.Icon = strPtrToType(data["icon"])
	plan.Description = strPtrToType(data["description"])
	plan.IPAccess = expandStringValues(data["ip_access"])
	plan.EnforceTFA = boolPtrToType(data["enforce_tfa"])
	plan.AdminAccess = boolPtrToType(data["admin_access"])
	plan.AppAccess = boolPtrToType(data["app_access"])

	// plan.Permissions = expandInt64Values(data["permissions"])
	// plan.Users = expandStringValues(data["users"])
	// plan.Roles = expandStringValues(data["roles"])

	switch v := resp.(type) {
	case *resource.CreateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.ReadResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.UpdateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	}
}

func boolPtrToType(any any) types.Bool {
	if b, ok := any.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

// utility funcs
func expandInt64List(list []types.Int64) []int64 {
	result := []int64{}
	for _, v := range list {
		if !v.IsNull() {
			result = append(result, v.ValueInt64())
		}
	}
	return result
}

func expandStringList(list []types.String) []string {
	result := []string{}
	for _, v := range list {
		if !v.IsNull() {
			result = append(result, v.ValueString())
		}
	}
	return result
}

func expandInt64Values(v any) []types.Int64 {
	arr := []types.Int64{}
	if list, ok := v.([]any); ok {
		for _, i := range list {
			if f, ok := i.(float64); ok {
				arr = append(arr, types.Int64Value(int64(f)))
			}
		}
	}
	return arr
}

func expandStringValues(v any) []types.String {
	arr := []types.String{}
	if list, ok := v.([]any); ok {
		for _, i := range list {
			if s, ok := i.(string); ok {
				arr = append(arr, types.StringValue(s))
			}
		}
	}
	return arr
}
