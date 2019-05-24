package gandi

import (
	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	h := meta.(hosting.Hosting)
	region := hosting.Region{
		ID: d.Get("region_id").(string),
	}
	image, err := h.ImageByName(d.Get("name").(string), region)
	if err != nil {
		return err
	}
	d.SetId(image.ID)
	d.Set("image_id", image.ID)
	d.Set("disk_id", image.DiskID)
	d.Set("size", int(image.Size))
	return nil
}
