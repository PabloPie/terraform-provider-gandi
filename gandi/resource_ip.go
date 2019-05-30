package gandi

import (
	"fmt"
	"log"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPCreate,
		Read:   resourceIPRead,
		Update: resourceIPUpdate,
		Delete: resourceIPDelete,
		Exists: resourceIPExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v != 4 && v != 6 {
						errs = append(errs, fmt.Errorf("%q must be either 4 or 6, got: %d", key, v))
					}
					return
				},
				ForceNew: true,
			},
			// Computed
			"ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vm_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIPCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	region := hosting.Region{
		ID: d.Get("region_id").(string),
	}
	version := d.Get("version").(int)
	ip, err := h.CreateIP(region, hosting.IPVersion(version))
	if err != nil {
		return err
	}

	d.SetId(ip.ID)
	return resourceIPRead(d, m)
}

func resourceIPRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	ipfilter := hosting.IPFilter{
		ID: d.Id(),
	}
	ips, err := h.ListIPs(ipfilter)
	if err != nil {
		return err
	}
	if len(ips) < 1 {
		d.SetId("")
		log.Printf("[ERR] IP with ID %s not found", d.Id())
		return nil
	}
	ip := ips[0]
	d.Set("state", ip.State)
	d.Set("region_id", ip.RegionID)
	d.Set("ip", ip.IP)
	d.Set("version", int(ip.Version))
	d.Set("vm_id", ip.VM)
	return nil
}

// IPs are immutable, updating means deleting and recreating the resource
// If deletion fails, the new IP is still created and the ressource updated,
// the old IP has to be manually deleted.
// Add a parameter to modify this behaviour? update_on_failure?
func resourceIPUpdate(d *schema.ResourceData, m interface{}) error {
	if !d.HasChange("region_id") && !d.HasChange("version") {
		return resourceIPRead(d, m)
	}

	var err error
	h := m.(hosting.Hosting)
	regionid := hosting.Region{ID: d.Get("region_id").(string)}
	ipversion := hosting.IPVersion(d.Get("version").(int))
	ip := hosting.IPAddress{ID: d.Id()}

	if err := h.DeleteIP(ip); err != nil {
		log.Printf("[WARN] Error deleting IP %s: %s", ip.ID, err)
	}
	if ip, err = h.CreateIP(regionid, ipversion); err != nil {
		return err
	}
	d.SetId(ip.ID)
	return resourceIPRead(d, m)
}

func resourceIPDelete(d *schema.ResourceData, m interface{}) (err error) {
	h := m.(hosting.Hosting)
	ip := hosting.IPAddress{
		ID: d.Id(),
	}
	if exists, _ := resourceIPExists(d, m); exists {
		err = h.DeleteIP(ip)
	}
	return err
}

func resourceIPExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)
	ips, err := h.ListIPs(hosting.IPFilter{ID: d.Id()})
	if len(ips) > 0 && err == nil {
		return true, nil
	}
	return false, err
}
