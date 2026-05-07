package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	providerpkg "github.com/rossvideo/terraform-provider-st2138/internal/provider"
)

func main() {
	providerserver.Serve(context.Background(), providerpkg.New, providerserver.ServeOpts{
		// Full provider address for OpenTofu registry publishing.
		Address: "registry.opentofu.org/rossvideo/st2138",
	})
}
