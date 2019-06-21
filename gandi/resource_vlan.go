package gandi

import (
	"log"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceVlanCreate,
		Read:   resourceVlanRead,
		Update: resourceVlanUpdate,
		Delete: resourceVlanDelete,
		Exists: resourceVlanExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Size in GB",
			},
			"gateway": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVlanCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vlanspec := hosting.VlanSpec{
		RegionID: d.Get("region_id").(string),
	}
	if name, ok := d.GetOk("name"); ok {
		vlanspec.Name = name.(string)
	}
	if subnet, ok := d.GetOk("subnet"); ok {
		vlanspec.Subnet = subnet.(string)
	}
	vlan, err := h.CreateVlan(vlanspec)
	if err != nil {
		return err
	}
	d.SetId(vlan.ID)
	return resourceVlanRead(d, m)
}

func resourceVlanRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vlanfilter := hosting.VlanFilter{
		ID: []string{d.Id()},
	}
	vlans, err := h.ListVlans(vlanfilter)
	if err != nil {
		return err
	}
	if len(vlans) < 1 {
		log.Printf("[ERR] Vlan with ID %s not found", d.Id())
		d.SetId("")
		return nil
	}
	vlan := vlans[0]
	d.Set("region_id", vlan.RegionID)
	d.Set("name", vlan.Name)
	d.Set("gateway", vlan.Gateway)
	d.Set("subnet", vlan.Subnet)
	return nil
}

func resourceVlanUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	h := m.(hosting.Hosting)
	vlan := hosting.Vlan{ID: d.Id()}
	if d.HasChange("name") {
		_, newname := d.GetChange("name")
		revlan, err := h.RenameVlan(vlan, newname.(string))
		if err != nil {
			return err
		}
		d.Set("name", revlan.Name)
		d.SetPartial("name")
	}
	if d.HasChange("gateway") {
		_, newGateway := d.GetChange("gateway")
		revlan, err := h.UpdateVlanGW(vlan, newGateway.(string))
		if err != nil {
			return err
		}
		d.Set("gateway", revlan.Gateway)
		d.SetPartial("gateway")
	}
	d.Partial(false)
	return resourceVlanRead(d, m)
}

func resourceVlanDelete(d *schema.ResourceData, m interface{}) (err error) {
	h := m.(hosting.Hosting)
	if exists, _ := resourceVlanExists(d, m); exists {
		vlan := hosting.Vlan{
			ID: d.Id(),
		}
		err = h.DeleteVlan(vlan)
	}
	return
}

func resourceVlanExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)
	vlan, err := h.VlanFromName(d.Get("name").(string))
	return vlan.ID != "", err
}
