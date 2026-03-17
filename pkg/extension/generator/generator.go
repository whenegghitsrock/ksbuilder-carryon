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
		deps = append(deps, &chart.Dependency{
			Name: "backend",
			Tags: []string{"agent"},
		})
	}

	type extMeta struct {
		APIVersion       string                 `yaml:"apiVersion"`
		Name             string                 `yaml:"name"`
		Version          string                 `yaml:"version"`
		DisplayName      map[string]string      `yaml:"displayName"`
		Description      map[string]string      `yaml:"description"`
		Category         string                 `yaml:"category"`
		Keywords         []string               `yaml:"keywords"`
		Home             string                 `yaml:"home"`
		Sources          []string               `yaml:"sources"`
		KubeVersion      string                 `yaml:"kubeVersion"`
		KSVersion        string                 `yaml:"ksVersion"`
		Maintainers      []*chart.Maintainer    `yaml:"maintainers"`
		Provider         map[string]interface{} `yaml:"provider"`
		Icon             string                 `yaml:"icon"`
		Screenshots      []string               `yaml:"screenshots"`
		Dependencies     []*chart.Dependency    `yaml:"dependencies"`
		InstallationMode string                 `yaml:"installationMode"`
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
		InstallationMode: "HostOnly",
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
	tpl := `frontend:
  enabled: {{ .Frontend }}
  image:
    repository: kubespheredev/{{ .Name }}-frontend
    tag: latest

backend:
  enabled: {{ .Backend }}
  image:
    repository: kubespheredev/{{ .Name }}-api
    tag: latest
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
	return b.Bytes(), nil
}
