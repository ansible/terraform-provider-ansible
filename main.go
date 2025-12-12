package main

import (
	"context"
	"log"

	"github.com/ansible/terraform-provider-ansible/framework"
	"github.com/ansible/terraform-provider-ansible/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	ctx := context.Background()
	primary := provider.Provider()
	providers := []func() tfprotov5.ProviderServer{
		func() tfprotov5.ProviderServer {
			return schema.NewGRPCProviderServer(primary)
		},
		providerserver.NewProtocol5(framework.New(primary)),
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt

	err = tf5server.Serve("registry.terraform.io/ansible/ansible", muxServer.ProviderServer, serveOpts...)
	if err != nil {
		log.Fatal(err)
	}
}
