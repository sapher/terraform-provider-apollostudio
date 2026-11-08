package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/sapher/terraform-provider-apollostudio/pkg/client"
)

const (
	// The provider reads its configuration from the APOLLO_KEY and APOLLO_ORG_ID environment variables.
	providerConfig = `
		provider "apollostudio" {}
	`

	testAccVariantName  = "current"
	testAccSubgraphName = "countries"
	testAccSubgraphUrl  = "https://countries.example.com/graphql"
	testAccSchema       = "type Query {\n  hello: String\n}"
	testAccSchemaV2     = "type Query {\n  hello: String\n  world: String\n}"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"apollostudio": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast with a clear message when the acceptance test credentials are missing.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, name := range []string{"APOLLO_KEY", "APOLLO_ORG_ID"} {
		if os.Getenv(name) == "" {
			t.Fatalf("%s must be set for acceptance tests", name)
		}
	}
}

// testAccSkipUnlessAcc skips tests that call the Apollo API outside resource.Test when TF_ACC is not set.
func testAccSkipUnlessAcc(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests are skipped unless TF_ACC is set")
	}
	testAccPreCheck(t)
}

func testClient() *client.ApolloClient {
	host := "https://graphql.api.apollographql.com/api/graphql"
	return client.NewClient(host, os.Getenv("APOLLO_KEY"), os.Getenv("APOLLO_ORG_ID"))
}

// testAccGraphId returns a unique graph ID so that tests (and the Terraform version matrix) never collide.
func testAccGraphId() string {
	return "tfacc-" + uuid.New().String()[0:8]
}

// testAccGraphConfig declares a throwaway graph named after its ID.
func testAccGraphConfig(graphId string) string {
	return fmt.Sprintf(`
resource "apollostudio_graph" "fixture" {
  id          = %[1]q
  name        = %[1]q
  description = "graph dedicated for terraform provider acceptance testing"
}
`, graphId)
}

// testAccSubgraphConfig declares a throwaway graph with one subgraph published to the "current" variant.
func testAccSubgraphConfig(graphId string, schema string) string {
	return testAccGraphConfig(graphId) + fmt.Sprintf(`
resource "apollostudio_subgraph" "fixture" {
  graph_id     = apollostudio_graph.fixture.id
  variant_name = %q
  name         = %q
  url          = %q
  schema       = %q
}
`, testAccVariantName, testAccSubgraphName, testAccSubgraphUrl, schema)
}
