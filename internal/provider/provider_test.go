package provider

import (
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

const (
	// Environment variables used to configure the provider.
	providerConfig = `
		provider "apollostudio" {}
	`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"apollostudio": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testClient() *client.ApolloClient {
	host := "https://graphql.api.apollographql.com/api/graphql"
	apiKey := os.Getenv("APOLLO_KEY")
	orgId := os.Getenv("APOLLO_ORG_ID")
	return client.NewClient(host, apiKey, orgId)
}
