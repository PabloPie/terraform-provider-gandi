package gandi

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"gandi": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if apiKey := os.Getenv("GANDI_API_KEY"); apiKey == "" {
		t.Fatal("GANDI_API_KEY must be set for acceptance tests")
	}
}
