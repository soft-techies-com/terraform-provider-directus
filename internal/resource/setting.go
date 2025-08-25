package resource

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/soft-techies-com/terraform-provider-directus/internal/client"
)

// SettingResource implements the directus_setting resource
type SettingResource struct{ client *client.Directus }

// SettingModel represents the setting resource model
type SettingModel struct {
	ID                            types.String `tfsdk:"id"` // Directus settings endpoint returns a singleton id
	ProjectName                   types.String `tfsdk:"project_name"`
	ProjectURL                    types.String `tfsdk:"project_url"`
	ProjectColor                  types.String `tfsdk:"project_color"`
	ProjectLogo                   types.String `tfsdk:"project_logo"`
	PublicForeground              types.String `tfsdk:"public_foreground"`
	PublicBackground              types.String `tfsdk:"public_background"`
	PublicNote                    types.String `tfsdk:"public_note"`
	AuthLoginAttempts             types.Int64  `tfsdk:"auth_login_attempts"`
	AuthPasswordPolicy            types.String `tfsdk:"auth_password_policy"`
	StorageAssetTransform         types.String `tfsdk:"storage_asset_transform"`
	StorageAssetPresets           types.String `tfsdk:"storage_asset_presets"`
	CustomCSS                     types.String `tfsdk:"custom_css"`
	StorageDefaultFolder          types.String `tfsdk:"storage_default_folder"`
	Basemaps                      types.String `tfsdk:"basemaps"`
	MapboxKey                     types.String `tfsdk:"mapbox_key"`
	ModuleBar                     types.String `tfsdk:"module_bar"`
	ProjectDescriptor             types.String `tfsdk:"project_descriptor"`
	DefaultLanguage               types.String `tfsdk:"default_language"`
	CustomAspectRatios            types.String `tfsdk:"custom_aspect_ratios"`
	PublicFavicon                 types.String `tfsdk:"public_favicon"`
	DefaultAppearance             types.String `tfsdk:"default_appearance"`
	DefaultThemeLight             types.String `tfsdk:"default_theme_light"`
	ThemeLightOverrides           types.String `tfsdk:"theme_light_overrides"`
	DefaultThemeDark              types.String `tfsdk:"default_theme_dark"`
	ThemeDarkOverrides            types.String `tfsdk:"theme_dark_overrides"`
	ReportErrorURL                types.String `tfsdk:"report_error_url"`
	ReportBugURL                  types.String `tfsdk:"report_bug_url"`
	ReportFeatureURL              types.String `tfsdk:"report_feature_url"`
	PublicRegistration            types.Bool   `tfsdk:"public_registration"`
	PublicRegistrationVerifyEmail types.Bool   `tfsdk:"public_registration_verify_email"`
	PublicRegistrationRole        types.String `tfsdk:"public_registration_role"`
	PublicRegistrationEmailFilter types.String `tfsdk:"public_registration_email_filter"`
	VisualEditorURLs              types.String `tfsdk:"visual_editor_urls"`
	AcceptedTerms                 types.Bool   `tfsdk:"accepted_terms"`
	ProjectID                     types.String `tfsdk:"project_id"`
}

// NewSettingResource returns a new setting resource
func NewSettingResource() resource.Resource { return &SettingResource{} }

// Metadata returns the resource type name
func (r *SettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting"
}

// Schema defines the schema for the resource
func (r *SettingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"id":                               rschema.StringAttribute{Computed: true},
			"project_name":                     rschema.StringAttribute{Optional: true},
			"project_url":                      rschema.StringAttribute{Optional: true},
			"project_color":                    rschema.StringAttribute{Optional: true},
			"project_logo":                     rschema.StringAttribute{Optional: true},
			"public_foreground":                rschema.StringAttribute{Optional: true},
			"public_background":                rschema.StringAttribute{Optional: true},
			"public_note":                      rschema.StringAttribute{Optional: true},
			"auth_login_attempts":              rschema.Int64Attribute{Optional: true},
			"auth_password_policy":             rschema.StringAttribute{Optional: true},
			"storage_asset_transform":          rschema.StringAttribute{Optional: true},
			"storage_asset_presets":            rschema.StringAttribute{Optional: true},
			"custom_css":                       rschema.StringAttribute{Optional: true},
			"storage_default_folder":           rschema.StringAttribute{Optional: true},
			"basemaps":                         rschema.StringAttribute{Optional: true},
			"mapbox_key":                       rschema.StringAttribute{Optional: true},
			"module_bar":                       rschema.StringAttribute{Optional: true},
			"project_descriptor":               rschema.StringAttribute{Optional: true},
			"default_language":                 rschema.StringAttribute{Optional: true},
			"custom_aspect_ratios":             rschema.StringAttribute{Optional: true},
			"public_favicon":                   rschema.StringAttribute{Optional: true},
			"default_appearance":               rschema.StringAttribute{Optional: true},
			"default_theme_light":              rschema.StringAttribute{Optional: true},
			"theme_light_overrides":            rschema.StringAttribute{Optional: true},
			"default_theme_dark":               rschema.StringAttribute{Optional: true},
			"theme_dark_overrides":             rschema.StringAttribute{Optional: true},
			"report_error_url":                 rschema.StringAttribute{Optional: true},
			"report_bug_url":                   rschema.StringAttribute{Optional: true},
			"report_feature_url":               rschema.StringAttribute{Optional: true},
			"public_registration":              rschema.BoolAttribute{Optional: true},
			"public_registration_verify_email": rschema.BoolAttribute{Optional: true},
			"public_registration_role":         rschema.StringAttribute{Optional: true},
			"public_registration_email_filter": rschema.StringAttribute{Optional: true},
			"visual_editor_urls":               rschema.StringAttribute{Optional: true},
			"accepted_terms":                   rschema.BoolAttribute{Optional: true},
			"project_id":                       rschema.StringAttribute{Computed: true},
		},
	}
}

// Configure configures the resource
func (r *SettingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Directus)
}

// Create creates a new setting
func (r *SettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Directus settings is a singleton; PATCH /settings to update and then GET to read id
	var plan SettingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{}
	if v := plan.ProjectName.ValueString(); v != "" {
		payload["project_name"] = v
	}
	if v := plan.ProjectURL.ValueString(); v != "" {
		payload["project_url"] = v
	}
	if v := plan.ProjectColor.ValueString(); v != "" {
		payload["project_color"] = v
	}
	if v := plan.ProjectLogo.ValueString(); v != "" {
		payload["project_logo"] = v
	}
	if v := plan.PublicForeground.ValueString(); v != "" {
		payload["public_foreground"] = v
	}
	if v := plan.PublicBackground.ValueString(); v != "" {
		payload["public_background"] = v
	}
	if v := plan.PublicNote.ValueString(); v != "" {
		payload["public_note"] = v
	}
	if !plan.AuthLoginAttempts.IsNull() {
		payload["auth_login_attempts"] = plan.AuthLoginAttempts.ValueInt64()
	}
	if v := plan.AuthPasswordPolicy.ValueString(); v != "" {
		payload["auth_password_policy"] = v
	}
	if v := plan.StorageAssetTransform.ValueString(); v != "" {
		payload["storage_asset_transform"] = v
	}
	if v := plan.StorageAssetPresets.ValueString(); v != "" {
		payload["storage_asset_presets"] = v
	}
	if v := plan.CustomCSS.ValueString(); v != "" {
		payload["custom_css"] = v
	}
	if v := plan.StorageDefaultFolder.ValueString(); v != "" {
		payload["storage_default_folder"] = v
	}
	if v := plan.Basemaps.ValueString(); v != "" {
		payload["basemaps"] = v
	}
	if v := plan.MapboxKey.ValueString(); v != "" {
		payload["mapbox_key"] = v
	}
	if v := plan.ModuleBar.ValueString(); v != "" {
		payload["module_bar"] = v
	}
	if v := plan.ProjectDescriptor.ValueString(); v != "" {
		payload["project_descriptor"] = v
	}
	if v := plan.DefaultLanguage.ValueString(); v != "" {
		payload["default_language"] = v
	}
	if v := plan.CustomAspectRatios.ValueString(); v != "" {
		payload["custom_aspect_ratios"] = v
	}
	if v := plan.PublicFavicon.ValueString(); v != "" {
		payload["public_favicon"] = v
	}
	if v := plan.DefaultAppearance.ValueString(); v != "" {
		payload["default_appearance"] = v
	}
	if v := plan.DefaultThemeLight.ValueString(); v != "" {
		payload["default_theme_light"] = v
	}
	if v := plan.ThemeLightOverrides.ValueString(); v != "" {
		payload["theme_light_overrides"] = v
	}
	if v := plan.DefaultThemeDark.ValueString(); v != "" {
		payload["default_theme_dark"] = v
	}
	if v := plan.ThemeDarkOverrides.ValueString(); v != "" {
		payload["theme_dark_overrides"] = v
	}
	if v := plan.ReportErrorURL.ValueString(); v != "" {
		payload["report_error_url"] = v
	}
	if v := plan.ReportBugURL.ValueString(); v != "" {
		payload["report_bug_url"] = v
	}
	if v := plan.ReportFeatureURL.ValueString(); v != "" {
		payload["report_feature_url"] = v
	}
	if !plan.PublicRegistration.IsNull() {
		payload["public_registration"] = plan.PublicRegistration.ValueBool()
	}
	if !plan.PublicRegistrationVerifyEmail.IsNull() {
		payload["public_registration_verify_email"] = plan.PublicRegistrationVerifyEmail.ValueBool()
	}
	if v := plan.PublicRegistrationRole.ValueString(); v != "" {
		payload["public_registration_role"] = v
	}
	if v := plan.PublicRegistrationEmailFilter.ValueString(); v != "" {
		payload["public_registration_email_filter"] = v
	}
	if v := plan.VisualEditorURLs.ValueString(); v != "" {
		payload["visual_editor_urls"] = v
	}
	if !plan.AcceptedTerms.IsNull() {
		payload["accepted_terms"] = plan.AcceptedTerms.ValueBool()
	}

	httpResp, err := r.client.Request(ctx, http.MethodPatch, "/settings", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, nil); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}

	// read back to capture id and normalized values
	r.readIntoState(ctx, &plan, resp)
}

// Read reads the setting
func (r *SettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SettingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := state
	r.readIntoState(ctx, &plan, resp)
}

func (r *SettingResource) readIntoState(ctx context.Context, plan *SettingModel, resp interface{}) {
	// helper compatible with both CreateResponse and ReadResponse via small interface
	type diagLike interface{ AddError(string, string) }
	var d diagLike
	switch v := resp.(type) {
	case *resource.CreateResponse:
		d = &v.Diagnostics
	case *resource.ReadResponse:
		d = &v.Diagnostics
	case *resource.UpdateResponse:
		d = &v.Diagnostics
	default:
		return
	}
	apiResp := struct {
		Data map[string]any `json:"data"`
	}{}
	httpResp, err := r.client.Request(ctx, http.MethodGet, "/settings", nil)
	if err != nil {
		d.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, &apiResp); err != nil {
		d.AddError("api error", err.Error())
		return
	}

	// Set all fields from the API response
	plan.ID = types.StringValue(str(apiResp.Data["id"]))
	plan.ProjectName = strPtrToType(apiResp.Data["project_name"])
	plan.ProjectURL = strPtrToType(apiResp.Data["project_url"])
	plan.ProjectColor = strPtrToType(apiResp.Data["project_color"])
	plan.ProjectLogo = strPtrToType(apiResp.Data["project_logo"])
	plan.PublicForeground = strPtrToType(apiResp.Data["public_foreground"])
	plan.PublicBackground = strPtrToType(apiResp.Data["public_background"])
	plan.PublicNote = strPtrToType(apiResp.Data["public_note"])

	// Handle numeric values
	if v, ok := apiResp.Data["auth_login_attempts"]; ok && v != nil {
		if intVal, ok := v.(float64); ok {
			plan.AuthLoginAttempts = types.Int64Value(int64(intVal))
		}
	}

	plan.AuthPasswordPolicy = strPtrToType(apiResp.Data["auth_password_policy"])
	plan.StorageAssetTransform = strPtrToType(apiResp.Data["storage_asset_transform"])
	plan.StorageAssetPresets = strPtrToType(apiResp.Data["storage_asset_presets"])
	plan.CustomCSS = strPtrToType(apiResp.Data["custom_css"])
	plan.StorageDefaultFolder = strPtrToType(apiResp.Data["storage_default_folder"])
	plan.Basemaps = strPtrToType(apiResp.Data["basemaps"])
	plan.MapboxKey = strPtrToType(apiResp.Data["mapbox_key"])
	plan.ModuleBar = strPtrToType(apiResp.Data["module_bar"])
	plan.ProjectDescriptor = strPtrToType(apiResp.Data["project_descriptor"])
	plan.DefaultLanguage = strPtrToType(apiResp.Data["default_language"])
	plan.CustomAspectRatios = strPtrToType(apiResp.Data["custom_aspect_ratios"])
	plan.PublicFavicon = strPtrToType(apiResp.Data["public_favicon"])
	plan.DefaultAppearance = strPtrToType(apiResp.Data["default_appearance"])
	plan.DefaultThemeLight = strPtrToType(apiResp.Data["default_theme_light"])
	plan.ThemeLightOverrides = strPtrToType(apiResp.Data["theme_light_overrides"])
	plan.DefaultThemeDark = strPtrToType(apiResp.Data["default_theme_dark"])
	plan.ThemeDarkOverrides = strPtrToType(apiResp.Data["theme_dark_overrides"])
	plan.ReportErrorURL = strPtrToType(apiResp.Data["report_error_url"])
	plan.ReportBugURL = strPtrToType(apiResp.Data["report_bug_url"])
	plan.ReportFeatureURL = strPtrToType(apiResp.Data["report_feature_url"])

	// Handle boolean values
	if v, ok := apiResp.Data["public_registration"]; ok && v != nil {
		if boolVal, ok := v.(bool); ok {
			plan.PublicRegistration = types.BoolValue(boolVal)
		}
	}

	if v, ok := apiResp.Data["public_registration_verify_email"]; ok && v != nil {
		if boolVal, ok := v.(bool); ok {
			plan.PublicRegistrationVerifyEmail = types.BoolValue(boolVal)
		}
	}

	plan.PublicRegistrationRole = strPtrToType(apiResp.Data["public_registration_role"])
	plan.PublicRegistrationEmailFilter = strPtrToType(apiResp.Data["public_registration_email_filter"])
	plan.VisualEditorURLs = strPtrToType(apiResp.Data["visual_editor_urls"])

	if v, ok := apiResp.Data["accepted_terms"]; ok && v != nil {
		if boolVal, ok := v.(bool); ok {
			plan.AcceptedTerms = types.BoolValue(boolVal)
		}
	}

	plan.ProjectID = strPtrToType(apiResp.Data["project_id"])

	switch v := resp.(type) {
	case *resource.CreateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.ReadResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	case *resource.UpdateResponse:
		v.Diagnostics.Append(v.State.Set(ctx, plan)...)
	}
}

// Update updates the setting
func (r *SettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SettingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := map[string]any{}
	if v := plan.ProjectName.ValueString(); v != "" {
		payload["project_name"] = v
	} else {
		payload["project_name"] = nil
	}
	if v := plan.ProjectURL.ValueString(); v != "" {
		payload["project_url"] = v
	} else {
		payload["project_url"] = nil
	}
	if v := plan.ProjectColor.ValueString(); v != "" {
		payload["project_color"] = v
	} else {
		payload["project_color"] = nil
	}
	if v := plan.ProjectLogo.ValueString(); v != "" {
		payload["project_logo"] = v
	} else {
		payload["project_logo"] = nil
	}
	if v := plan.PublicForeground.ValueString(); v != "" {
		payload["public_foreground"] = v
	} else {
		payload["public_foreground"] = nil
	}
	if v := plan.PublicBackground.ValueString(); v != "" {
		payload["public_background"] = v
	} else {
		payload["public_background"] = nil
	}
	if v := plan.PublicNote.ValueString(); v != "" {
		payload["public_note"] = v
	} else {
		payload["public_note"] = nil
	}
	if !plan.AuthLoginAttempts.IsNull() {
		payload["auth_login_attempts"] = plan.AuthLoginAttempts.ValueInt64()
	} else {
		payload["auth_login_attempts"] = nil
	}
	if v := plan.AuthPasswordPolicy.ValueString(); v != "" {
		payload["auth_password_policy"] = v
	} else {
		payload["auth_password_policy"] = nil
	}
	if v := plan.StorageAssetTransform.ValueString(); v != "" {
		payload["storage_asset_transform"] = v
	} else {
		payload["storage_asset_transform"] = nil
	}
	if v := plan.StorageAssetPresets.ValueString(); v != "" {
		payload["storage_asset_presets"] = v
	} else {
		payload["storage_asset_presets"] = nil
	}
	if v := plan.CustomCSS.ValueString(); v != "" {
		payload["custom_css"] = v
	} else {
		payload["custom_css"] = nil
	}
	if v := plan.StorageDefaultFolder.ValueString(); v != "" {
		payload["storage_default_folder"] = v
	} else {
		payload["storage_default_folder"] = nil
	}
	if v := plan.Basemaps.ValueString(); v != "" {
		payload["basemaps"] = v
	} else {
		payload["basemaps"] = nil
	}
	if v := plan.MapboxKey.ValueString(); v != "" {
		payload["mapbox_key"] = v
	} else {
		payload["mapbox_key"] = nil
	}
	if v := plan.ModuleBar.ValueString(); v != "" {
		payload["module_bar"] = v
	} else {
		payload["module_bar"] = nil
	}
	if v := plan.ProjectDescriptor.ValueString(); v != "" {
		payload["project_descriptor"] = v
	} else {
		payload["project_descriptor"] = nil
	}
	if v := plan.DefaultLanguage.ValueString(); v != "" {
		payload["default_language"] = v
	} else {
		payload["default_language"] = nil
	}
	if v := plan.CustomAspectRatios.ValueString(); v != "" {
		payload["custom_aspect_ratios"] = v
	} else {
		payload["custom_aspect_ratios"] = nil
	}
	if v := plan.PublicFavicon.ValueString(); v != "" {
		payload["public_favicon"] = v
	} else {
		payload["public_favicon"] = nil
	}
	if v := plan.DefaultAppearance.ValueString(); v != "" {
		payload["default_appearance"] = v
	} else {
		payload["default_appearance"] = nil
	}
	if v := plan.DefaultThemeLight.ValueString(); v != "" {
		payload["default_theme_light"] = v
	} else {
		payload["default_theme_light"] = nil
	}
	if v := plan.ThemeLightOverrides.ValueString(); v != "" {
		payload["theme_light_overrides"] = v
	} else {
		payload["theme_light_overrides"] = nil
	}
	if v := plan.DefaultThemeDark.ValueString(); v != "" {
		payload["default_theme_dark"] = v
	} else {
		payload["default_theme_dark"] = nil
	}
	if v := plan.ThemeDarkOverrides.ValueString(); v != "" {
		payload["theme_dark_overrides"] = v
	} else {
		payload["theme_dark_overrides"] = nil
	}
	if v := plan.ReportErrorURL.ValueString(); v != "" {
		payload["report_error_url"] = v
	} else {
		payload["report_error_url"] = nil
	}
	if v := plan.ReportBugURL.ValueString(); v != "" {
		payload["report_bug_url"] = v
	} else {
		payload["report_bug_url"] = nil
	}
	if v := plan.ReportFeatureURL.ValueString(); v != "" {
		payload["report_feature_url"] = v
	} else {
		payload["report_feature_url"] = nil
	}
	if !plan.PublicRegistration.IsNull() {
		payload["public_registration"] = plan.PublicRegistration.ValueBool()
	} else {
		payload["public_registration"] = nil
	}
	if !plan.PublicRegistrationVerifyEmail.IsNull() {
		payload["public_registration_verify_email"] = plan.PublicRegistrationVerifyEmail.ValueBool()
	} else {
		payload["public_registration_verify_email"] = nil
	}
	if v := plan.PublicRegistrationRole.ValueString(); v != "" {
		payload["public_registration_role"] = v
	} else {
		payload["public_registration_role"] = nil
	}
	if v := plan.PublicRegistrationEmailFilter.ValueString(); v != "" {
		payload["public_registration_email_filter"] = v
	} else {
		payload["public_registration_email_filter"] = nil
	}
	if v := plan.VisualEditorURLs.ValueString(); v != "" {
		payload["visual_editor_urls"] = v
	} else {
		payload["visual_editor_urls"] = nil
	}
	if !plan.AcceptedTerms.IsNull() {
		payload["accepted_terms"] = plan.AcceptedTerms.ValueBool()
	} else {
		payload["accepted_terms"] = nil
	}

	httpResp, err := r.client.Request(ctx, http.MethodPatch, "/settings", payload)
	if err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	if err := parseResp(httpResp, nil); err != nil {
		resp.Diagnostics.AddError("api error", err.Error())
		return
	}
	r.readIntoState(ctx, &plan, resp)
}

// Delete deletes the setting
func (r *SettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Not supported: settings is singleton; we interpret delete as a no-op and remove from state
	resp.State.RemoveResource(ctx)
}
