package gandi

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGandiDisk_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiDiskBasic,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiDiskExists("gandi_disk.accTestDisk"),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "size", "10"),
				),
			},
		},
	})
}

func TestAccGandiDisk_withNameandSize(t *testing.T) {
	diskname := fmt.Sprintf("gandidisk%d", acctest.RandIntRange(0, 10000))
	disksize := acctest.RandIntRange(5, 30)
	diskconfig := fmt.Sprintf(testAccGandiDiskNameAndSize, diskname, disksize)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiImage + diskconfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiDiskExists("gandi_disk.accTestDisk"),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "size", strconv.Itoa(disksize)),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "name", diskname),
				),
			},
		},
	})
}

func TestAccGandiDisk_withNameFromImage(t *testing.T) {
	diskname := fmt.Sprintf("gandidisk%d", acctest.RandIntRange(0, 10000))
	diskconfig := fmt.Sprintf(testAccGandiDiskWithNameFromImage, diskname)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiImage + diskconfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiDiskExists("gandi_disk.accTestDisk"),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "size", "3"),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "name", diskname),
				),
			},
		},
	})
}

func TestAccGandiDisk_withNameAndSizeFromImage(t *testing.T) {
	diskname := fmt.Sprintf("gandidisk%d", acctest.RandIntRange(0, 10000))
	disksize := acctest.RandIntRange(5, 30)
	diskconfig := fmt.Sprintf(testAccGandiDiskWithNameAndSizeFromImage, diskname, disksize)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGandiRegion + testAccGandiImage + diskconfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckGandiDiskExists("gandi_disk.accTestDisk"),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "size", strconv.Itoa(disksize)),
					resource.TestCheckResourceAttr("gandi_disk.accTestDisk", "name", diskname),
				),
			},
		},
	})
}

func testCheckGandiDiskExists(disk string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[disk]
		if !ok {
			return fmt.Errorf("Not found: %s", disk)
		}
		diskid := rs.Primary.ID
		h := testAccProvider.Meta().(hosting.Hosting)
		disks, err := h.ListDisks(hosting.DiskFilter{ID: diskid})
		if err != nil {
			return err
		}
		if len(disks) < 1 {
			return fmt.Errorf("Error: Disk %q does not exist", rs.Primary.ID)
		}
		return nil
	}
}

var testAccGandiDiskFromImage = `
resource "gandi_disk" "accTestDisk" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	src_disk_id = "${data.gandi_image.accTestImage.disk_id}"
}
`

var testAccGandiDiskWithNameFromImage = `
resource "gandi_disk" "accTestDisk" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	src_disk_id = "${data.gandi_image.accTestImage.disk_id}"
	name = "%s"
}
`

var testAccGandiDiskWithNameAndSizeFromImage = `
resource "gandi_disk" "accTestDisk" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	src_disk_id = "${data.gandi_image.accTestImage.disk_id}"
	name = "%s"
	size = "%d"
}
`

var testAccGandiDiskNameAndSize = `
resource "gandi_disk" "accTestDisk" {
	region_id = "${data.gandi_region.accTestRegion.id}"
	name = "%s"
	size = "%d"
}
`

var testAccGandiDiskBasic = `
resource "gandi_disk" "accTestDisk" {
	region_id = "${data.gandi_region.accTestRegion.id}"
}
`
