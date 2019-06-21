package gandi

import (
	"fmt"
	"testing"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGandiIP_ipv6(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiIPv6,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiIPExists("gandi_ip.accTestIP"),
				),
			},
		},
	})
}

func TestAccGandiIP_ipv4(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiIPv4,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiIPExists("gandi_ip.accTestIP"),
				),
			},
		},
	})
}

func testCheckGandiIPExists(ip string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[ip]
		if !ok {
			return fmt.Errorf("Not found: %s", ip)
		}
		ipid := rs.Primary.ID
		h := testAccProvider.Meta().(hosting.Hosting)
		ips, err := h.ListIPs(hosting.IPFilter{ID: ipid})
		if err != nil {
			return err
		}
		if len(ips) < 1 {
			return fmt.Errorf("Error: IP %q does not exist", ipid)
		}
		return nil
	}
}

var testAccGandiIPv6 = `
resource "gandi_ip" "accTestIP" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	version = 6
}
`

var testAccGandiIPv4 = `
resource "gandi_ip" "accTestIP" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	version = 6
}
`
