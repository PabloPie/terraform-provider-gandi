package gandi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/terraform"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestGandi_ipDiff(t *testing.T) {
	cases := []struct {
		oldips         []hosting.IPAddress
		newips         []hosting.IPAddress
		expectedDetach []hosting.IPAddress
		expectedAttach []hosting.IPAddress
	}{
		// Detach everything
		{
			oldips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			newips:         []hosting.IPAddress{},
			expectedAttach: []hosting.IPAddress{},
			expectedDetach: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
		},
		// Attach everything
		{
			oldips: []hosting.IPAddress{},
			newips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			expectedAttach: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
		},
		// Same IPS, nothing to do
		{
			oldips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			newips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			// attach is never nil, while detach can be nil
			expectedAttach: []hosting.IPAddress{},
		},
		// Add 1 ip
		{
			oldips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			newips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
			},
			expectedAttach: []hosting.IPAddress{
				{ID: "3"},
			},
		},
		// Add 2 ip, detach 2 ips
		{
			oldips: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
			newips: []hosting.IPAddress{
				{ID: "3"},
				{ID: "4"},
			},
			expectedAttach: []hosting.IPAddress{
				{ID: "3"},
				{ID: "4"},
			},
			expectedDetach: []hosting.IPAddress{
				{ID: "1"},
				{ID: "2"},
			},
		},
	}

	for i, set := range cases {
		todetach, toattach := ipDiff(set.oldips, set.newips)
		if !reflect.DeepEqual(toattach, set.expectedAttach) {
			t.Fatalf("Error in case %d with attached interfaces, expected %#v, got %#v instead", i, set.expectedAttach, toattach)
		}
		if !reflect.DeepEqual(todetach, set.expectedDetach) {
			t.Fatalf("Error in case %d with detached interfaces, expected %#v, got %#v instead", i, set.expectedDetach, todetach)
		}
	}
}

func TestAccGandiVM_basic(t *testing.T) {
	machinename := fmt.Sprintf("gandivm-%d", acctest.RandInt())
	config := fmt.Sprintf(testAccGandiVM_basic, machinename)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiVMExists(machinename),
				),
			},
		},
	})
}

func testCheckGandiVMExists(vm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[vm]
		if !ok {
			return fmt.Errorf("Not found: %s", vm)
		}
		h := testAccProvider.Meta().(hosting.Hosting)
		vm, err := h.VMFromName(vm)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		if vm.ID == "" {
			return fmt.Errorf("Error: Machine %q does not exist", rs.Primary.ID)
		}
		return nil
	}
}

var testAccGandiVM_basic = `
data "gandi_region" "datacenter" {
	region_code = "FR-SD6"
}

data "gandi_image" "debian" {
	name = "Debian 9"
	region_id = "${data.gandi_region.datacenter.id}"
}

resource "gandi_ip" "testip" {
	region_id = "${data.gandi_region.datacenter.id}"
	version = 6
}

resource "gandi_disk" "testsystemdisk" {
	region_id = "${data.gandi_region.datacenter.id}"
	src_disk_id = "${data.gandi_image.debian.disk_id}"
	name = "acctest_sysdisk"
	size = 10
}

resource "gandi_ssh" "testkey" {
	name = "acctest_key"
	value = "ssh-rsa asdfasdf pasdfa"
}

resource "gandi_vm" "test" {
	region_id = "${data.gandi_region.datacenter.id}"
	name = "%s"
	ips {
		id = "${gandi_ip.testip.id}"
	}
	boot_disk {
	  name = "${gandi_disk.testsystemdisk.name}"
	}
	ssh_keys = ["acctest_key"]
  }
`
