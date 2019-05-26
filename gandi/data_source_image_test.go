package gandi

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccGandiImage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiImage,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.gandi_image.accTestImage", "id", "407"),
					resource.TestCheckResourceAttr("data.gandi_image.accTestImage", "disk_id", "21548621"),
				),
			},
		},
	})
}

var testAccGandiImage = `
data "gandi_image" "accTestImage" {
	name = "Debian 9"
	region_id = "${data.gandi_region.accTestRegion.id}"
}
`
