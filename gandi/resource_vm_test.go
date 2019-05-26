package gandi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestGandi_testContains(t *testing.T) {
	cases := []struct {
		iplist []interface{}
		ip     hosting.IPAddress
		exists bool
	}{
		{
			iplist: []interface{}{
				map[string]interface{}{
					"id": "1",
				},
			},
			ip: hosting.IPAddress{
				ID: "1",
			},
			exists: true,
		},
		{
			iplist: []interface{}{
				map[string]interface{}{
					"id": "1",
				},
			},
			ip: hosting.IPAddress{
				ID: "2",
			},
			exists: false,
		},
		{
			iplist: []interface{}{
				map[string]interface{}{
					"id": "1",
				},
				map[string]interface{}{
					"id": "5",
				},
			},
			ip: hosting.IPAddress{
				ID: "5",
			},
			exists: true,
		},
	}
	for i, set := range cases {
		contained := containsIP(set.iplist, set.ip)
		if set.exists != contained {
			t.Fatalf("Error in case %d, expected %t, got %t instead", i, set.exists, contained)
		}
	}
}

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
	vmname := fmt.Sprintf("gandivm-%d", acctest.RandIntRange(1, 1000))
	vmconfig := fmt.Sprintf(testAccGandiVMbasic, vmname)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: vmconfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiVMExists("gandi_vm.accTestVM"),
					resource.TestCheckResourceAttr("gandi_vm.accTestVM", "name", vmname),
					resource.TestCheckResourceAttr("gandi_vm.accTestVM", "cores", "1"),
					resource.TestCheckResourceAttr("gandi_vm.accTestVM", "memory", "512"),
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
		vmid := rs.Primary.ID
		vms, err := h.DescribeVM(hosting.VMFilter{ID: vmid})
		if err != nil {
			return err
		}
		if len(vms) < 1 {
			return fmt.Errorf("Error: VM %q does not exist", vmid)
		}
		return nil
	}
}

var testAccGandiVMbasic = `
data "gandi_region" "datacenter" {
    region_code = "FR-SD6"
}

data "gandi_image" "debian9" {
  name = "Debian 9"
  region_id = "${data.gandi_region.datacenter.id}"
}

resource "gandi_ip" "ip1" {
  region_id = "${data.gandi_region.datacenter.id}"
  version = 6
}

resource "gandi_disk" "systemdisk1" {
  region_id = "${data.gandi_region.datacenter.id}"
  src_disk_id = "${data.gandi_image.debian9.disk_id}"
}

resource "gandi_vm" "accTestVM" {
  region_id = "${data.gandi_region.datacenter.id}"
  name = "%s"
  ips {
    id = "${gandi_ip.ip1.id}"
  }
  boot_disk {
    name = "${gandi_disk.systemdisk1.name}"
  }
  userpass {
    login = "testlogin"
    password = "Passwordfortest123!"
  }
}
`
