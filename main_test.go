package main

import (
	"context"
	"errors"
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func TestMainServesRegisteredProvider(t *testing.T) {
	original := serveProvider
	t.Cleanup(func() { serveProvider = original })
	called := false
	serveProvider = func(ctx context.Context, factory func() frameworkprovider.Provider, options providerserver.ServeOpts) error {
		called = true
		if ctx == nil {
			t.Error("Serve context is nil")
		}
		provider := factory()
		metadata := &frameworkprovider.MetadataResponse{}
		provider.Metadata(ctx, frameworkprovider.MetadataRequest{}, metadata)
		if metadata.TypeName != "st2138" {
			t.Errorf("provider type name = %q, want st2138", metadata.TypeName)
		}
		if options.Address != "registry.opentofu.org/rossvideo/st2138" {
			t.Errorf("server address = %q", options.Address)
		}
		return errors.New("test server stopped")
	}

	main()
	if !called {
		t.Fatal("main() did not serve the provider")
	}
}
