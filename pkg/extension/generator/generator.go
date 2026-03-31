package generator

import (
	"bytes"
	"text/template"

	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

// ExtensionYAML generates extension.yaml from spec (standard mode only).
func ExtensionYAML(s *spec.Spec) ([]byte, error) {
	if s.Mode != spec.ModeStandard {
		return nil, nil // app/simple use their own templates for now
	}

	displayName := s.Metadata.DisplayName
	if displayName == nil {
		displayName = map[string]string{"zh": s.Name, "en": s.Name}
	}
	if displayName["zh"] == "" {
		displayName["zh"] = s.Name
	}
	if displayName["en"] == "" {
		displayName["en"] = s.Name
	}
	desc := s.Metadata.Description
	if desc == nil {
		desc = map[string]string{
			"zh": "这是一个示例扩展组件，这是它的描述",
			"en": "This is a sample extension, and this is its description",
		}
	}
	if desc["zh"] == "" {
		desc["zh"] = "这是一个示例扩展组件，这是它的描述"
	}
	if desc["en"] == "" {
		desc["en"] = "This is a sample extension, and this is its description"
	}

	keywords := s.Metadata.Keywords
	if len(keywords) == 0 {
		keywords = []string{s.Name, s.Metadata.Category}
	}

	var deps []*chart.Dependency
	if s.HasFrontend() {
		deps = append(deps, &chart.Dependency{
			Name: "frontend",
			Tags: []string{"extension"},
		})
	}
	if s.HasBackend() {
		backendDep := &chart.Dependency{Name: "backend"}
		// Backend-only extensions don't need the "agent" tag; keep it for mixed frontend+backend mode.
		if s.HasFrontend() {
			backendDep.Tags = []string{"agent"}
		}
		deps = append(deps, backendDep)
	}

	// sigs.k8s.io/yaml uses json tags (not yaml tags) for marshaling; use json tags for lowercase keys
	type extMeta struct {
		APIVersion       string                 `json:"apiVersion"`
		Name             string                 `json:"name"`
		Version          string                 `json:"version"`
		DisplayName      map[string]string      `json:"displayName"`
		Description      map[string]string      `json:"description"`
		Category         string                 `json:"category"`
		Keywords         []string               `json:"keywords"`
		Home             string                 `json:"home"`
		Sources          []string               `json:"sources"`
		KubeVersion      string                 `json:"kubeVersion"`
		KSVersion        string                 `json:"ksVersion"`
		Maintainers      []*chart.Maintainer    `json:"maintainers"`
		Provider         map[string]interface{} `json:"provider"`
		Icon             string                 `json:"icon"`
		Screenshots      []string               `json:"screenshots"`
		Dependencies     []*chart.Dependency    `json:"dependencies"`
		InstallationMode string                 `json:"installationMode"`
	}

	// HostOnly: extension subcharts stay on the host cluster only.
	// Multicluster: only when both frontend+backend are present.
	instMode := "HostOnly"
	if s.HasFrontend() && s.HasBackend() {
		instMode = "Multicluster"
	}

	m := extMeta{
		APIVersion:       "kubesphere.io/v1alpha1",
		Name:             s.Name,
		Version:          s.Version,
		DisplayName:      displayName,
		Description:      desc,
		Category:         s.Metadata.Category,
		Keywords:         keywords,
		Home:             "https://kubesphere.io",
		Sources:          []string{"https://github.com/kubesphere"},
		KubeVersion:      ">=1.19.0-0",
		KSVersion:        ">=4.0.0-0",
		Icon:             "./static/favicon.svg",
		Screenshots:      []string{"./static/screenshots/screenshot.png"},
		Dependencies:     deps,
		InstallationMode: instMode,
	}

	if s.Metadata.Author != "" {
		m.Maintainers = []*chart.Maintainer{{
			Name:  s.Metadata.Author,
			Email: s.Metadata.Email,
			URL:   s.Metadata.URL,
		}}
		m.Provider = map[string]interface{}{
			"zh": map[string]string{
				"name":  s.Metadata.Author,
				"email": s.Metadata.Email,
				"url":   s.Metadata.URL,
			},
			"en": map[string]string{
				"name":  s.Metadata.Author,
				"email": s.Metadata.Email,
				"url":   s.Metadata.URL,
			},
		}
	} else {
		m.Maintainers = []*chart.Maintainer{{Name: "admin"}}
		m.Provider = map[string]interface{}{
			"zh": map[string]string{"name": "admin", "email": "", "url": ""},
			"en": map[string]string{"name": "admin", "email": "", "url": ""},
		}
	}

	return yaml.Marshal(m)
}

// ExtensionYAMLForChartMode generates extension.yaml for app and simple modes (from existing helm chart).
// Uses fixed maintainer "demoauthor", fixed category "ai-machine-learning", no dependencies.
func ExtensionYAMLForChartMode(name string) ([]byte, error) {
	displayName := map[string]string{"zh": name, "en": name}
	desc := map[string]string{
		"zh": "这是一个示例扩展组件，这是它的描述",
		"en": "This is a sample extension, and this is its description",
	}
	type extMeta struct {
		APIVersion       string                 `json:"apiVersion"`
		Name             string                 `json:"name"`
		Version          string                 `json:"version"`
		DisplayName      map[string]string      `json:"displayName"`
		Description      map[string]string      `json:"description"`
		Category         string                 `json:"category"`
		Keywords         []string               `json:"keywords"`
		Home             string                 `json:"home"`
		Sources          []string               `json:"sources"`
		KubeVersion      string                 `json:"kubeVersion"`
		KSVersion        string                 `json:"ksVersion"`
		Maintainers      []*chart.Maintainer    `json:"maintainers"`
		Provider         map[string]interface{} `json:"provider"`
		Icon             string                 `json:"icon"`
		Screenshots      []string               `json:"screenshots"`
		InstallationMode string                 `json:"installationMode"`
	}
	m := extMeta{
		APIVersion:       "kubesphere.io/v1alpha1",
		Name:             name,
		Version:          "0.1.0",
		DisplayName:      displayName,
		Description:      desc,
		Category:         "ai-machine-learning",
		Keywords:         []string{name, "ai-machine-learning"},
		Home:             "https://kubesphere.io",
		Sources:          []string{"https://github.com/kubesphere"},
		KubeVersion:      ">=1.19.0-0",
		KSVersion:        ">=4.0.0-0",
		Icon:             "./static/favicon.svg",
		Screenshots:      []string{"./static/screenshots/screenshot.png"},
		Maintainers:      []*chart.Maintainer{{Name: "demoauthor", Email: "", URL: ""}},
		Provider: map[string]interface{}{
			"zh": map[string]string{"name": "demoauthor", "email": "", "url": ""},
			"en": map[string]string{"name": "demoauthor", "email": "", "url": ""},
		},
		InstallationMode: "HostOnly",
	}
	return yaml.Marshal(m)
}

// PermissionsYAML generates permissions.yaml from spec.
func PermissionsYAML(s *spec.Spec) ([]byte, error) {
	switch s.Permissions {
	case spec.PermApp:
		return permissionsAppYAML()
	case spec.PermSimple:
		return permissionsSimpleYAML()
	default:
		return permissionsDefaultYAML()
	}
}

func permissionsDefaultYAML() ([]byte, error) {
	return []byte(`kind: ClusterRole
rules:
  - verbs:
      - '*'
    apiGroups:
      - 'extensions.kubesphere.io'
    resources:
      - '*'

---
kind: Role
rules:
  - verbs:
      - '*'
    apiGroups:
      - ''
      - 'apps'
      - 'batch'
      - 'app.k8s.io'
      - 'autoscaling'
    resources:
      - '*'
  - verbs:
      - '*'
    apiGroups:
      - 'networking.k8s.io'
    resources:
      - 'ingresses'
      - 'networkpolicies'
`), nil
}

func permissionsAppYAML() ([]byte, error) {
	return []byte(`kind: ClusterRole
rules:
  - verbs:
      - '*'
    apiGroups:
      - 'extensions.kubesphere.io'
      - 'application.kubesphere.io'
    resources:
      - '*'
  - verbs:
      - '*'
    apiGroups:
      - ''
    resources:
      - 'configmaps'
      - 'namespaces'
  - verbs:
      - '*'
    apiGroups:
      - 'rbac.authorization.k8s.io'
    resources:
      - '*'

---
kind: Role
rules:
  - verbs:
      - '*'
    apiGroups:
      - ''
      - 'apps'
      - 'batch'
      - 'app.k8s.io'
      - 'autoscaling'
    resources:
      - '*'
  - verbs:
      - '*'
    apiGroups:
      - 'networking.k8s.io'
    resources:
      - 'ingresses'
      - 'networkpolicies'
`), nil
}

func permissionsSimpleYAML() ([]byte, error) {
	return []byte(`kind: ClusterRole
rules:
  - verbs:
      - '*'
    apiGroups:
      - 'extensions.kubesphere.io'
      - 'jsbundles.extensions.kubesphere.io'
    resources:
      - '*'
  - verbs:
      - '*'
    apiGroups:
      - ''
    resources:
      - 'configmaps'
  - verbs:
      - '*'
    apiGroups:
      - 'rbac.authorization.k8s.io'
    resources:
      - 'clusterroles'
---
kind: Role
rules:
  - verbs:
      - '*'
    apiGroups:
      - ''
      - 'apps'
      - 'batch'
      - 'app.k8s.io'
      - 'autoscaling'
    resources:
      - '*'
  - verbs:
      - '*'
    apiGroups:
      - 'networking.k8s.io'
    resources:
      - 'ingresses'
      - 'networkpolicies'
`), nil
}

// ValuesYAML generates values.yaml from spec (standard mode only).
func ValuesYAML(s *spec.Spec) ([]byte, error) {
	if s.Mode != spec.ModeStandard {
		return nil, nil
	}
	tpl := `{{- if .Frontend }}
frontend:
  enabled: {{ .Frontend }}
  image:
    repository: kubespheredev/{{ .Name }}-frontend
    tag: latest
    pullPolicy: IfNotPresent
{{- end -}}
{{- if and .Frontend .Backend }}
{{ "" }}
{{- end -}}
{{- if .Backend }}
backend:
  enabled: {{ .Backend }}
  image:
    repository: kubespheredev/{{ .Name }}-api
    tag: latest
    pullPolicy: IfNotPresent
{{- end }}
`
	t, err := template.New("values").Parse(tpl)
	if err != nil {
		return nil, err
	}
	data := struct {
		Name     string
		Frontend bool
		Backend  bool
	}{s.Name, s.HasFrontend(), s.HasBackend()}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, err
	}
	out := bytes.TrimLeft(b.Bytes(), "\n")
	return out, nil
}
