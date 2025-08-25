package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/directus-terraform/internal/client"
)

var _ resource.ResourceWithImportState = &PermissionResource{}

// PermissionResource implements the directus_permission resource
type PermissionResource struct{ client *client.Directus }

// PermissionModel represents the permission resource model
type PermissionModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Collection  types.String `tfsdk:"collection"`
	Action      types.String `tfsdk:"action"`      // read, create, update, delete
	Permissions types.String `tfsdk:"permissions"` // JSON string
	Validation  types.String `tfsdk:"validation"`  // JSON string
	Fields      types.List   `tfsdk:"fields"`      // list of strings
	Policy      types.String `tfsdk:"policy"`
	Presets     types.String `tfsdk:"presets"` // JSON string
	System      types.Bool   `tfsdk:"system"`
}

// NewPermissionResource returns a new permission resource
func NewPermissionResource() resource.Resource { return &PermissionResource{} }

// Metadata returns the resource type name
func (r *PermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

// Schema defines the schema for the resource
func (r *PermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":          rschema.Int64Attribute{Computed: true},
			"collection":  rschema.StringAttribute{Required: true},
			"action":      rschema.StringAttribute{Required: true, Description: "one of read, create, update, delete"},
			"permissions": rschema.StringAttribute{Optional: true, Description: "JSON string of permissions filter"},
			"validation":  rschema.StringAttribute{Optional: true, Description: "JSON string of validation rules"},
			"presets":     rschema.StringAttribute{Optional: true, Description: "JSON string of presets"},
			"fields":      rschema.ListAttribute{Optional: true, ElementType: types.StringType},
			"policy":      rschema.StringAttribute{Required: true},
			"system":      rschema.BoolAttribute{Optional: true, Computed: true},
		},
	}
}

// Configure configures the resource
func (r *PermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Directus)
}

// Create creates a new permission
func (r *PermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"collection": plan.Collection.ValueString(),
		"action":     plan.Action.ValueString(),
		"policy":     plan.Policy.ValueString(),
	}

	if s := plan.Permissions.ValueString(); s != "" {
		payload["permissions"] = jsonRaw(s)
	}
	if s := plan.Validation.ValueString(); s != "" {
		payload["validation"] = jsonRaw(s)
	}
	if s := plan.Presets.ValueString(); s != "" {
		payload["presets"] = jsonRaw(s)
	}
	if !plan.Fields.IsNull() {
		var fields []string
		plan.Fields.ElementsAs(ctx, &fields, false)
		payload["fields"] = fields
	}

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	httpResp, err := r.client.Request(ctx, http.MethodPost, "/permissions", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, &apiResp); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	// handle ID robustly
	if v, ok := apiResp.Data["id"].(float64); ok {
		plan.ID = types.Int64Value(int64(v))
	} else if v, ok := apiResp.Data["id"].(int); ok {
		plan.ID = types.Int64Value(int64(v))
	} else if v, ok := apiResp.Data["id"].(int64); ok {
		plan.ID = types.Int64Value(v)
	} else if v, ok := apiResp.Data["id"].(string); ok {
		if idInt, err := strconv.ParseInt(v, 10, 64); err == nil {
			plan.ID = types.Int64Value(idInt)
		}
	}
	if v, ok := apiResp.Data["policy"].(string); ok {
		plan.Policy = types.StringValue(v)
	} else {
		plan.Policy = types.StringValue("")
	}
	if v, ok := apiResp.Data["system"].(bool); ok {
		plan.System = types.BoolValue(v)
	} else {
		plan.System = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the permission
func (r *PermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	httpResp, err := r.client.Request(
		ctx,
		http.MethodGet,
		"/permissions/"+strconv.FormatInt(state.ID.ValueInt64(), 10),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, &apiResp); err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	state.Collection = types.StringValue(str(apiResp.Data["collection"]))
	state.Action = types.StringValue(str(apiResp.Data["action"]))

	// fields
	if v, ok := apiResp.Data["fields"].([]any); ok {
		elems := make([]attr.Value, 0, len(v))
		for _, f := range v {
			elems = append(elems, types.StringValue(str(f)))
		}
		state.Fields, _ = types.ListValue(types.StringType, elems)
	}

	// id
	if v, ok := apiResp.Data["id"].(float64); ok {
		state.ID = types.Int64Value(int64(v))
	}
	// json blobs
	if v, ok := apiResp.Data["permissions"]; ok && v != nil {
		b, _ := json.Marshal(v)
		state.Permissions = types.StringValue(string(b))
	}
	if v, ok := apiResp.Data["validation"]; ok && v != nil {
		b, _ := json.Marshal(v)
		state.Validation = types.StringValue(string(b))
	}
	if v, ok := apiResp.Data["presets"]; ok && v != nil {
		b, _ := json.Marshal(v)
		state.Presets = types.StringValue(string(b))
	}

	if v, ok := apiResp.Data["policy"].(string); ok {
		state.Policy = types.StringValue(v)
	} else {
		state.Policy = types.StringValue("")
	}
	if v, ok := apiResp.Data["system"].(bool); ok {
		state.System = types.BoolValue(v)
	} else {
		state.System = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan PermissionModel
    var state PermissionModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Preserve ID from current state
    plan.ID = state.ID

    payload := map[string]any{
        "collection": plan.Collection.ValueString(),
        "action":     plan.Action.ValueString(),
        "policy":     plan.Policy.ValueString(),
    }

    if s := plan.Permissions.ValueString(); s != "" {
        payload["permissions"] = jsonRaw(s)
    } else {
        payload["permissions"] = nil
    }
    if s := plan.Validation.ValueString(); s != "" {
        payload["validation"] = jsonRaw(s)
    } else {
        payload["validation"] = nil
    }
    if s := plan.Presets.ValueString(); s != "" {
        payload["presets"] = jsonRaw(s)
    } else {
        payload["presets"] = nil
    }
    if !plan.Fields.IsNull() {
        var fields []string
        plan.Fields.ElementsAs(ctx, &fields, false)
        payload["fields"] = fields
    } else {
        payload["fields"] = nil
    }

    // Send PATCH
    httpResp, err := r.client.Request(
        ctx,
        http.MethodPatch,
        "/permissions/"+strconv.FormatInt(plan.ID.ValueInt64(), 10),
        payload,
    )
    if err != nil {
        resp.Diagnostics.AddError("api error", err.Error())
        return
    }
    if err := parseResp(httpResp, nil); err != nil {
        resp.Diagnostics.AddError("api error", err.Error())
        return
    }

    // ✅ Re-read resource from API to refresh id/system/etc
    readResp, err := r.client.Request(
        ctx,
        http.MethodGet,
        "/permissions/"+strconv.FormatInt(plan.ID.ValueInt64(), 10),
        nil,
    )
    if err != nil {
        resp.Diagnostics.AddError("api error", err.Error())
        return
    }

    apiResp := struct {
        Data map[string]any `json:"data"`
    }{}
    if err := parseResp(readResp, &apiResp); err != nil {
        resp.Diagnostics.AddError("api error", err.Error())
        return
    }

    // reuse same parsing logic from Read()
    var newState PermissionModel
    newState.ID = plan.ID
    newState.Collection = types.StringValue(str(apiResp.Data["collection"]))
    newState.Action = types.StringValue(str(apiResp.Data["action"]))

    if v, ok := apiResp.Data["fields"].([]any); ok {
        elems := make([]attr.Value, 0, len(v))
        for _, f := range v {
            elems = append(elems, types.StringValue(str(f)))
        }
        newState.Fields, _ = types.ListValue(types.StringType, elems)
    }

    if v, ok := apiResp.Data["permissions"]; ok && v != nil {
        b, _ := json.Marshal(v)
        newState.Permissions = types.StringValue(string(b))
    }
    if v, ok := apiResp.Data["validation"]; ok && v != nil {
        b, _ := json.Marshal(v)
        newState.Validation = types.StringValue(string(b))
    }
    if v, ok := apiResp.Data["presets"]; ok && v != nil {
        b, _ := json.Marshal(v)
        newState.Presets = types.StringValue(string(b))
    }

    if v, ok := apiResp.Data["policy"].(string); ok {
        newState.Policy = types.StringValue(v)
    } else {
        newState.Policy = types.StringValue("")
    }
    if v, ok := apiResp.Data["system"].(bool); ok {
        newState.System = types.BoolValue(v)
    } else {
        newState.System = types.BoolValue(false)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}



// Delete deletes the permission
func (r *PermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PermissionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Request(
		ctx,
		http.MethodDelete,
		"/permissions/"+strconv.FormatInt(state.ID.ValueInt64(), 10),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, nil); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
}

// ImportState allows terraform import support for directus_permission
func (r *PermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idStr := req.ID

	if _, err := strconv.Atoi(idStr); err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Permission ID must be an integer, got: %q", idStr),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idStr)...)
}
