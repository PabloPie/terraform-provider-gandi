package gandi

import (
	"fmt"

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
			},
			"src_disk_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt, // In GB
				Optional: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 {
						errs = append(errs, fmt.Errorf("%q must positive (>0GB), got: %d", key, v))
					}
					return
				},
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
	srcdisk, fromImage := d.GetOk("src_disk_id")
	var disk hosting.Disk
	var err error
	if fromImage {
		diskimage := hosting.DiskImage{
			DiskID:   srcdisk.(string),
			RegionID: d.Get("region_id").(string),
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
