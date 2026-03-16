package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/extension"
	"github.com/kubesphere/ksbuilder/pkg/extension/spec"
)

type inputPromptContent struct {
	text     string
	optional bool
	errorMsg string
}

type selectPromptContent struct {
	text              string
	items             []string
	startInSearchMode bool
}

type createOptions struct {
	from            string
	typ             string // standard | app | simple
	templateTypeIdx int    // 0=standard, 1=frontend, 2=backend (set by prompt when from scratch)
}

type Category struct {
	DisplayNameEN  string
	NormalizedName string
}

var Categories = []Category{
	{
		DisplayNameEN:  "AI / LLM",
		NormalizedName: "ai-machine-learning",
	},
	{
		DisplayNameEN:  "DeepSeek",
		NormalizedName: "deepseek",
	},
	{
		DisplayNameEN:  "Database",
		NormalizedName: "database",
	},
	{
		DisplayNameEN:  "Observability",
		NormalizedName: "observability",
	},
	{
		DisplayNameEN:  "CI / CD",
		NormalizedName: "integration-delivery",
	},
	{
		DisplayNameEN:  "Networking",
		NormalizedName: "networking",
	},
	{
		DisplayNameEN:  "Security",
		NormalizedName: "security",
	},
	{
		DisplayNameEN:  "Storage",
		NormalizedName: "storage",
	},
	{
		DisplayNameEN:  "Streaming and messaging",
		NormalizedName: "streaming-messaging",
	},
	{
		DisplayNameEN:  "Computing",
		NormalizedName: "computing",
	},
	{
		DisplayNameEN:  "DevTools",
		NormalizedName: "dev-tools",
	},
}

func getCategoryDisplayNames(categories []Category) []string {
	var names []string
	for _, c := range categories {
		names = append(names, c.DisplayNameEN)
	}
	return names
}

// overlayTypeFromTemplateIndex maps templateTypeIdx (0=standard, 1=frontend, 2=backend) to a label.
func overlayTypeFromTemplateIndex(idx int) string {
	switch idx {
	case 1:
		return "frontend"
	case 2:
		return "backend"
	default:
		return "standard"
	}
}

func createExtensionCmd() *cobra.Command {
	o := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new KubeSphere extension",
		Long:  "Create a new extension. Interactive mode prompts for type (standard/app/simple). Use --type=standard|app|simple with --from=<chart> for app/simple to skip prompts.",
		Args:  cobra.ExactArgs(0),
		RunE:  o.run,
	}
	cmd.Flags().StringVar(&o.from, "from", "", "application helm chart file path of application class")
	cmd.Flags().StringVar(&o.typ, "type", "standard", "extension type: standard (default), app, or simple. app/simple require --from")

	return cmd
}

func (o *createOptions) run(c *cobra.Command, _ []string) error {
	typ := o.typ
	if typ == "" {
		typ = "standard"
	}
	// Two-step interactive selection when --type not set
	if !c.Flags().Changed("type") {
		// Step 1: from scratch or from chart
		sourcePrompt := selectPromptContent{
			text:  "Create from",
			items: []string{"From scratch (Standard template with frontend+backend)", "From existing Helm chart (.tgz)"},
		}
		sourceIdx := promptGetSelect(sourcePrompt)
		if sourceIdx == 1 {
			// Step 2: App store or Simple
			chartTypePrompt := selectPromptContent{
				text:  "Chart type",
				items: []string{"App store (Application CR, for marketplace)", "Simple (extract as subchart, link/iframe)"},
			}
			chartTypeIdx := promptGetSelect(chartTypePrompt)
			if chartTypeIdx == 0 {
				typ = "app"
			} else {
				typ = "simple"
			}
		} else {
			// sourceIdx==0: from scratch, prompt template type
			templateTypePrompt := selectPromptContent{
				text:  "Template type",
				items: []string{"Standard (frontend+backend)", "Frontend only", "Backend only"},
			}
			o.templateTypeIdx = promptGetSelect(templateTypePrompt)
		}
	}

	switch typ {
	case "app":
		from := o.from
		if from == "" {
			fromPrompt := inputPromptContent{
				text:     "Chart file path (e.g. ./demo-0.1.0.tgz)",
				errorMsg: "Chart path can't be empty",
			}
			from = promptGetInput(fromPrompt)
		}
		return extension.CreateApp(from)
	case "simple":
		from := o.from
		if from == "" {
			fromPrompt := inputPromptContent{
				text:     "Chart file path (e.g. ./demo-0.1.0.tgz)",
				errorMsg: "Chart path can't be empty",
			}
			from = promptGetInput(fromPrompt)
		}
		return extension.CreateSimple(from)
	case "standard":
		// fall through to existing interactive flow below
	default:
		return fmt.Errorf("--type must be standard, app, or simple, got %q", typ)
	}

	// Standard interactive flow
	extensionNamePrompt := inputPromptContent{
		text:     "Please input extension name",
		errorMsg: "Extension name can't be empty",
	}
	name := promptGetInput(extensionNamePrompt)

	categoryDisplayNames := getCategoryDisplayNames(Categories)
	categoryPromptContent := selectPromptContent{
		text:  fmt.Sprintf("What category does %s belong to?", name),
		items: categoryDisplayNames,
	}
	categoryIdx := promptGetSelect(categoryPromptContent)

	authorPrompt := inputPromptContent{
		text:     "Please input extension author",
		errorMsg: "Extension author can't be empty",
	}
	author := promptGetInput(authorPrompt)

	emailPrompt := inputPromptContent{
		text:     "Please input Email",
		optional: true,
	}
	email := promptGetInput(emailPrompt)

	urlPrompt := inputPromptContent{
		text:     "Please input author's URL",
		optional: true,
	}
	url := promptGetInput(urlPrompt)

	pwd, _ := os.Getwd()
	p := path.Join(pwd, name)

	// Build declarative spec from prompts
	extSpec := &spec.Spec{
		Name:         name,
		Version:      "0.1.0",
		Mode:         spec.ModeStandard,
		Capabilities: spec.Capabilities{},
		Metadata: spec.Metadata{
			DisplayName: map[string]string{"zh": name, "en": name},
			Description: map[string]string{
				"zh": "这是一个示例扩展组件，这是它的描述",
				"en": "This is a sample extension, and this is its description",
			},
			Category: Categories[categoryIdx].NormalizedName,
			Keywords: []string{name, Categories[categoryIdx].NormalizedName},
			Author:    author,
			Email:     email,
			URL:       url,
		},
		Permissions: spec.PermDefault,
	}
	switch o.templateTypeIdx {
	case 1:
		extSpec.Capabilities = spec.Capabilities{Frontend: true, Backend: false}
	case 2:
		extSpec.Capabilities = spec.Capabilities{Frontend: false, Backend: true}
	default:
		extSpec.Capabilities = spec.Capabilities{Frontend: true, Backend: true}
	}
	if err := extension.CreateFromSpec(p, extSpec); err != nil {
		return err
	}

	if o.from != "" {
		chartPath := path.Join(pwd, o.from)
		appChart, err := os.ReadFile(chartPath)
		if err != nil {
			return err
		}
		if err = extension.CreateAppChart(p, name, appChart); err != nil {
			return err
		}
	}

	fmt.Printf("Directory: %s\n\n", p)
	fmt.Println("The extension charts has been created.")
	fmt.Println(createSuccessHint(true, extSpec.HasFrontend(), extSpec.HasBackend()))

	return nil
}

var (
	bold  = promptui.Styler(promptui.FGBold)
	faint = promptui.Styler(promptui.FGFaint)
)

func promptGetInput(pc inputPromptContent) string {
	prompt := promptui.Prompt{
		Label: pc.text,
	}

	if pc.optional {
		prompt.Templates = &promptui.PromptTemplates{
			Valid:   fmt.Sprintf("%s {{ . | bold }} %s ", bold(promptui.IconGood), bold("(optional):")),
			Success: fmt.Sprintf("{{ . | faint }} %s ", faint("(optional):")),
		}
	} else {
		prompt.Validate = func(input string) error {
			if len(strings.TrimSpace(input)) <= 0 {
				return errors.New(pc.errorMsg)
			}
			return nil
		}
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(result)
}

func promptGetSelect(pc selectPromptContent) int {
	prompt := promptui.Select{
		Label: pc.text,
		Items: pc.items,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(pc.items[index]), strings.ToLower(input))
		},
		StartInSearchMode: pc.startInSearchMode,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return idx
}
