package gandi

import (
	"fmt"
	"testing"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGandiSSH_basic(t *testing.T) {
	keyname := acctest.RandomWithPrefix("gandissh")
	keyvalue, _, _ := acctest.RandSSHKeyPair("")
	// api regex check requires a string after the key value
	keyconfig := fmt.Sprintf(testAccGandiSSHBasic, keyname, keyvalue+"nonexistentmail")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: keyconfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiSSHExists("gandi_ssh.accTestSSH"),
				),
			},
		},
	})
}

func testCheckGandiSSHExists(key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[key]
		if !ok {
			return fmt.Errorf("Not found: %s", key)
		}
		keyname := rs.Primary.Attributes["name"]
		h := testAccProvider.Meta().(hosting.Hosting)
		key := h.KeyFromName(keyname)

		if key.ID == "" {
			return fmt.Errorf("Error: SSH key %q does not exist", rs.Primary.ID)
		}
		return nil
	}
}

var testAccGandiSSHBasic = `
resource "gandi_ssh" "accTestSSH" {
	name = "%s"
	value = "%s"
}
`
