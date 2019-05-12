package gandi

import (
	"fmt"
	"regexp"

	"github.com/PabloPie/Gandi-Go/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiskCreate,
		Read:   resourceDiskRead,
		Update: resourceDiskUpdate,
		Delete: resourceDiskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
				// XXX: until we implement DC migration
				ForceNew: true,
			},
			"src_disk_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image"},
				Description:   "ID of the disk to use as source",
			},
			"image": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"src_disk_id"},
				Description:   "Name of the image to use as source",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: diskValidateName,
			},
			"size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Size in GB",
			},
			// Computed
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vm_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"boot_disk": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceDiskCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	diskspec := hosting.DiskSpec{
		RegionID: d.Get("region_id").(string),
	}
	if name, ok := d.GetOk("name"); ok {
		diskspec.Name = name.(string)
	}
	if size, ok := d.GetOk("size"); ok {
		diskspec.Size = size.(int)
	}
	srcdisk, fromDisk := d.GetOk("src_disk_id")
	image, fromImage := d.GetOk("image")
	var disk hosting.Disk
	var err error
	if fromDisk {
		diskimage := hosting.DiskImage{
			DiskID:   srcdisk.(string),
			RegionID: d.Get("region_id").(string),
		}
		if disk, err = h.CreateDiskFromImage(diskspec, diskimage); err != nil {
			return err
		}
	} else if fromImage {
		region := hosting.Region{
			ID: diskspec.RegionID,
		}
		diskimage, err := h.ImageByName(image.(string), region)
		if err != nil {
			return err
		}
		if disk, err = h.CreateDiskFromImage(diskspec, diskimage); err != nil {
			return err
		}
	} else {
		if disk, err = h.CreateDisk(diskspec); err != nil {
			return err
		}
	}
	d.SetId(disk.ID)
	return resourceDiskRead(d, m)
}

func resourceDiskRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	diskfilter := hosting.DiskFilter{
		ID: d.Id(),
	}
	disks, err := h.DescribeDisks(diskfilter)
	if err != nil {
		return err
	}
	if len(disks) < 1 {
		d.SetId("")
		return fmt.Errorf("Disk with ID %s not found", d.Id())
	}
	disk := disks[0]
	d.Set("state", disk.State)
	d.Set("region_id", disk.RegionID)
	d.Set("name", disk.Name)
	d.Set("size", disk.Size)
	d.Set("type", disk.Type)
	d.Set("vm_ids", disk.VM)
	d.Set("boot_disk", disk.BootDisk)
	return nil
}

func resourceDiskUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	h := m.(hosting.Hosting)
	disk := hosting.Disk{ID: d.Id(), Size: d.Get("size").(int)}
	if d.HasChange("name") {
		_, newname := d.GetChange("name")
		redisk, err := h.RenameDisk(disk, newname.(string))
		if err != nil {
			return err
		}
		d.Set("name", redisk.Name)
		d.SetPartial("name")
	}
	if d.HasChange("size") {
		oldsize, newsize := d.GetChange("size")
		if newsize.(int) <= oldsize.(int) {
			return fmt.Errorf("Disks cannot shrink in size")
		}
		// Extends doesnt change the size, it adds to it
		addedsize := newsize.(int) - oldsize.(int)
		exdisk, err := h.ExtendDisk(disk, uint(addedsize))
		if err != nil {
			return err
		}
		d.Set("size", exdisk.Size)
		d.SetPartial("size")
	}
	d.Partial(false)
	return resourceDiskRead(d, m)
}

func resourceDiskDelete(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	disk := hosting.Disk{
		ID: d.Id(),
	}
	err := h.DeleteDisk(disk)
	return err
}

func diskValidateName(value interface{}, name string) (warnings []string, errors []error) {
	r := regexp.MustCompile(`^[-_0-9a-z]{1,15}$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf("Invalid name: '%s', does not match %s", value.(string), r))
	}
	return
}
