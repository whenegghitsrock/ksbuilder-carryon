package cmd

import "strings"

func createSuccessHint(standardMode bool, hasFrontend, hasBackend bool, extName string) string {
	var b strings.Builder
	b.WriteString("Next steps:\n")
	if standardMode && (hasFrontend || hasBackend) {
		b.WriteString("  1. cd " + extName + "\n")
		b.WriteString("  2. make build-frontend build-backend\n")
		b.WriteString("     (or: REGISTRY=myreg NAMESPACE=myns make build-frontend build-backend)\n")
		b.WriteString("  3. make push\n")
		b.WriteString("     (or: REGISTRY=myreg NAMESPACE=myns make push)\n")
		b.WriteString("  4. Edit values.yaml: set frontend/backend image.repository\n")
		b.WriteString("  5. ksbuilder package .\n")
		b.WriteString("  6. ksbuilder publish . --target=cluster\n")
		if hasFrontend {
			b.WriteString("\nFor KubeSphere Console-style frontend (React, yarn dev):\n")
			b.WriteString("  See " + extName + "/frontend/README.md (links to dev-guide + extension-samples hello-world).\n")
			b.WriteString("  Optional: cd " + extName + " && mv frontend frontend-simple && yarn create ks-project frontend && cd frontend && yarn create:ext\n")
			b.WriteString("  (Console project in " + extName + "/frontend/, extension in frontend/extensions/. Run yarn dev from frontend/.)\n")
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
