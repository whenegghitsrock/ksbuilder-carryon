package cmd

import "strings"

func createSuccessHint(standardMode bool, hasFrontend, hasBackend bool) string {
	var b strings.Builder
	b.WriteString("Next steps:\n")
	if standardMode && (hasFrontend || hasBackend) {
		b.WriteString("  1. Build images:  make build-frontend build-backend\n")
		b.WriteString("  2. Push: make push (set REGISTRY, NAMESPACE)\n")
		b.WriteString("  3. Update values.yaml with your image repository.\n")
		b.WriteString("  4. Package:  ksbuilder package .\n")
		b.WriteString("  5. Publish: ksbuilder publish . --target=cluster\n")
		if hasFrontend {
			b.WriteString("\nFor KubeSphere Console-style frontend (React, yarn dev):\n")
			b.WriteString("  yarn create ks-project <project-name> && cd <project-name> && yarn create:ext\n")
		}
		return b.String()
	}
	b.WriteString("  cd into the new extension directory, then run `ksbuilder package .`.\n")
	return b.String()
}

func packageSuccessHint() string {
	return "Next: publish with `ksbuilder publish . --target=cluster` or `ksbuilder publish . --target=chartmuseum`."
}

func publishChartmuseumSuccessHint() string {
	return "Next: add the chart repo to Helm or install the chart from your repository."
}

func publishClusterSuccessHint() string {
	return "Next: open the KubeSphere console, search for this extension in Extensions or App Store, and install it."
}
