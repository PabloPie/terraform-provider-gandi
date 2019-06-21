package gandi

import (
	"log"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePrivateIP() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateIPCreate,
		Read:   resourcePrivateIPRead,
		Update: resourcePrivateIPUpdate,
		Delete: resourceIPDelete,
		Exists: resourcePrivateIPExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// XXX: pending iface migration implementation
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vlan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed
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

func resourcePrivateIPCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	region := hosting.Region{
		ID: d.Get("region_id").(string),
	}
	vlan := hosting.Vlan{
		ID:       d.Get("vlan_id").(string),
		RegionID: region.ID,
	}
	ip, err := h.CreatePrivateIP(vlan, d.Get("ip").(string))
	if err != nil {
		return err
	}

	d.SetId(ip.ID)
	return resourceIPRead(d, m)
}

func resourcePrivateIPRead(d *schema.ResourceData, m interface{}) error {
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
	d.Set("region_id", ip.RegionID)
	d.Set("state", ip.State)
	d.Set("ip", ip.IP)
	d.Set("vm_id", ip.VM)
	return nil
}

// Waiting for DC migration we delete the ip and recreate it manually on another Region
func resourcePrivateIPUpdate(d *schema.ResourceData, m interface{}) error {
	if !d.HasChange("region_id") {
		return resourceIPRead(d, m)
	}

	var err error
	h := m.(hosting.Hosting)
	region := hosting.Region{ID: d.Get("region_id").(string)}
	vlan := hosting.Vlan{
		ID:       d.Get("vlan_id").(string),
		RegionID: region.ID,
	}
	ipaddress := d.Get("ip").(string)
	ip := hosting.IPAddress{ID: d.Id()}

	if err := h.DeleteIP(ip); err != nil {
		log.Printf("[WARN] Error deleting IP %s: %s", ip.ID, err)
	}
	if ip, err = h.CreatePrivateIP(vlan, ipaddress); err != nil {
		return err
	}
	d.SetId(ip.ID)
	return resourceIPRead(d, m)
}

func resourcePrivateIPExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)
	ips, err := h.ListIPs(hosting.IPFilter{IP: d.Get("ip").(string)})
	return len(ips) > 0 && err == nil, err
}
