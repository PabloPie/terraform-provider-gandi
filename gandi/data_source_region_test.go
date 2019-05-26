package gandi

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccGandiRegion_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.gandi_region.accTestRegion", "id", "6"),
					resource.TestCheckResourceAttr("data.gandi_region.accTestRegion", "country", "France"),
				),
			},
		},
	})
}

var testAccGandiRegion = `
data "gandi_region" "accTestRegion" {
	region_code = "FR-SD6"
}
`
