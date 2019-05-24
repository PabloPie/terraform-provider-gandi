package gandi

import (
	"github.com/PabloPie/Gandi-Go/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRegionRead,
		Schema: map[string]*schema.Schema{
			"region_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed
			"country": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRegionRead(d *schema.ResourceData, meta interface{}) error {
	h := meta.(hosting.Hosting)
	region, err := h.RegionbyCode(d.Get("region_code").(string))
	if err != nil {
		return err
	}
	d.SetId(region.ID)
	d.Set("country", region.Country)
	return nil
}
