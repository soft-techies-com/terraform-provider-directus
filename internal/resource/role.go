package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/terraform-provider-directus/internal/client"
)

var _ resource.ResourceWithImportState = &RoleResource{}

// RoleResource implements the directus_role resource
type RoleResource struct{ client *client.Directus }

// RoleModel represents the role resource model
type RoleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Description types.String `tfsdk:"description"`
	// Parent      types.String `tfsdk:"parent"`
	// Children    types.List   `tfsdk:"children"`
	Policies types.Set `tfsdk:"policies"`
}

// NewRoleResource returns a new role resource
func NewRoleResource() resource.Resource { return &RoleResource{} }

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":          rschema.StringAttribute{Computed: true},
			"name":        rschema.StringAttribute{Required: true},
			"icon":        rschema.StringAttribute{Optional: true},
			"description": rschema.StringAttribute{Optional: true},
			// "parent":      rschema.StringAttribute{Optional: true},
			// "children":    rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"policies": rschema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Directus)
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create base role
	payload := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if !plan.Icon.IsNull() {
		payload["icon"] = plan.Icon.ValueString()
	}
	if !plan.Description.IsNull() {
		payload["description"] = plan.Description.ValueString()
	}
	// if !plan.Parent.IsNull() {
	// 	payload["parent"] = plan.Parent.ValueString()
	// }

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	httpResp, err := r.client.Request(ctx, http.MethodPost, "/roles", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error()+fmt.Sprintf(" payload: %v", payload))
		return
	}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	if v, ok := apiResp.Data["id"].(string); ok {
		plan.ID = types.StringValue(v)
	}

	// If policies were provided, PATCH them in
	if !plan.Policies.IsNull() {
		var policies []string
		if err := plan.Policies.ElementsAs(ctx, &policies, false); err == nil && len(policies) > 0 {
			createObjs := make([]map[string]any, 0, len(policies))
			for _, pid := range policies {
				createObjs = append(createObjs, map[string]any{
					"role":   plan.ID.ValueString(),
					"policy": pid, // FIX: just string id
				})
			}

			patchPayload := map[string]any{
				"policies": map[string]any{
					"create": createObjs,
					"update": []any{},
					"delete": []any{},
				},
			}

			httpResp, err := r.client.Request(ctx, http.MethodPatch, "/roles/"+plan.ID.ValueString(), patchPayload)
			if err != nil {
				resp.Diagnostics.AddError("api error", err.Error())
				return
			}
			if err := parseResp(httpResp, nil); err != nil {
				resp.Diagnostics.AddError("api error", err.Error())
				return
			}
		}
	}

	// Refresh from API to sync state
	r.refreshState(ctx, plan.ID.ValueString(), &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshState(ctx, state.ID.ValueString(), &state, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleModel
	var state RoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	// ---- 1. Update role fields ----
	payload := map[string]any{
		"name":        plan.Name.ValueString(),
		"icon":        nullableStr(plan.Icon),
		"description": nullableStr(plan.Description),
		// "parent":      nullableStr(plan.Parent),
	}

	httpResp, err := r.client.Request(ctx, http.MethodPatch, "/roles/"+plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, nil); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	// ---- 2. Handle policies ----
	if !plan.Policies.IsNull() {
		// desired from plan
		var desired []string
		if err := plan.Policies.ElementsAs(ctx, &desired, false); err != nil {
			resp.Diagnostics.AddError("plan error", fmt.Sprintf("failed to parse policies from plan: %w", err))
			return
		}

		// current from API
		apiResp := struct {
			Data map[string]any `json:"data"`
		}{}
		httpResp, err := r.client.Request(ctx, http.MethodGet, "/roles/"+plan.ID.ValueString()+"?fields=policies.policy.id", nil)
		if err != nil {
			resp.Diagnostics.AddError("api error", err.Error())
			return
		}
		if err := parseResp(httpResp, &apiResp); err != nil {
			resp.Diagnostics.AddError("api error", err.Error())
			return
		}

		var current []string
		if v, ok := apiResp.Data["policies"].([]any); ok {
			for _, p := range v {
				if m, ok := p.(map[string]any); ok {
					if policy, ok := m["policy"].(map[string]any); ok {
						current = append(current, str(policy["id"]))
					}
				}
			}
		}

		// build sets for diff
		currentSet := make(map[string]struct{})
		for _, id := range current {
			currentSet[id] = struct{}{}
		}
		desiredSet := make(map[string]struct{})
		for _, id := range desired {
			desiredSet[id] = struct{}{}
		}

		// determine create and delete
		var creates []map[string]any
		for _, id := range desired {
			if _, exists := currentSet[id]; !exists {
				creates = append(creates, map[string]any{
					"role":   plan.ID.ValueString(),
					"policy": map[string]any{"id": id},
				})
			}
		}

		var deletes []map[string]any
		for _, id := range current {
			if _, exists := desiredSet[id]; !exists {
				deletes = append(deletes, map[string]any{
					"role":   plan.ID.ValueString(),
					"policy": map[string]any{"id": id},
				})
			}
		}

		if len(creates) > 0 || len(deletes) > 0 {
			patchPayload := map[string]any{
				"policies": map[string]any{
					"create": creates,
					"update": []any{},
					"delete": deletes,
				},
			}

			httpResp, err := r.client.Request(ctx, http.MethodPatch, "/roles/"+plan.ID.ValueString(), patchPayload)
			if err != nil {
				resp.Diagnostics.AddError("api error", err.Error())
				return
			}
			if err := parseResp(httpResp, nil); err != nil {
				resp.Diagnostics.AddError("api error", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(ctx, http.MethodDelete, "/roles/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, nil); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idStr := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idStr)...)
}

// Helper functions

func (r *RoleResource) refreshState(ctx context.Context, id string, rm *RoleModel, diags diagCollector) bool {
	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	httpResp, err := r.client.Request(ctx, http.MethodGet, "/roles/"+id+"?fields=*,policies.policy.id", nil)
	if err != nil {
		diags.AddError("api error", err.Error())
		return false
	}
	if err := parseResp(httpResp, &apiResp); err != nil {
		if isNotFound(err) {
			return false
		}
		diags.AddError("api error", err.Error())
		return false
	}

	rm.Name = types.StringValue(str(apiResp.Data["name"]))
	rm.Icon = types.StringValue(str(apiResp.Data["icon"]))
	rm.Description = types.StringValue(str(apiResp.Data["description"]))
	// rm.Parent = types.StringValue(str(apiResp.Data["parent"]))

	if v, ok := apiResp.Data["policies"].([]any); ok {
		seen := make(map[string]struct{})
		elems := make([]attr.Value, 0, len(v))
		for _, p := range v {
			if m, ok := p.(map[string]any); ok {
				if policy, ok := m["policy"].(map[string]any); ok {
					id := str(policy["id"])
					if id != "" {
						if _, exists := seen[id]; !exists {
							seen[id] = struct{}{}
							elems = append(elems, types.StringValue(id))
						}
					}
				}
			}
		}

		rm.Policies, _ = types.SetValue(types.StringType, elems)
	} else {
		rm.Policies = types.SetNull(types.StringType)
	}

	return true
}

// diagCollector is a helper interface to unify diagnostics for refreshState
type diagCollector interface {
	AddError(summary string, detail string)
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "404")
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func boolVal(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case nil:
		return false
	default:
		return false
	}
}

func nullableStr(v types.String) any {
	if v.IsNull() || v.ValueString() == "" {
		return nil
	}
	return v.ValueString()
}

// jsonRaw stores a raw JSON string without double-encoding
type rawJSON string

func (r rawJSON) MarshalJSON() ([]byte, error) {
	if strings.TrimSpace(string(r)) == "" {
		return []byte("null"), nil
	}
	if !json.Valid([]byte(r)) {
		return nil, fmt.Errorf("invalid json string provided")
	}
	return []byte(r), nil
}

func jsonRaw(s string) rawJSON { return rawJSON(s) }

func strPtrToType(v any) types.String {
	if v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return types.StringNull()
		}
		return types.StringValue(s)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

func parseResp(resp *http.Response, out any) error {
	// defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var apiErr map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return fmt.Errorf("directus api %d: %v", resp.StatusCode, apiErr)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
