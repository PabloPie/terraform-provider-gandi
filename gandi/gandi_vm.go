package gandi

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// VM
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"farm": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Memory in MB",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v%64 != 0 {
						errs = append(errs, fmt.Errorf("%q must be a multiple of 64, got: %d", key, v))
					}
					return
				},
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			// Auth
			"ssh_keys": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        schema.TypeString,
				Description: "Names of the ssh keys allowed to connect",
			},
			"userpass": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"login": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			// Disks
			"boot_disk_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"boot_disk"},
				Description:   "ID of the disk to use as boot disk, must exist and be free",
			},
			"boot_disk": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"boot_disk_id"},
				Description:   "Disk spec of the disk used as boot disk",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// move boot disk id inside?
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							Computed: true,
						},
						"disk_image_id": {
							Type:     schema.TypeString,
							Required: true,
							Computed: true,
						},
						"os": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"disks": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			// IPs
			"ips": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ip_version": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v != 4 && v != 6 {
						errs = append(errs, fmt.Errorf("%q must be either 4 or 6, got: %d", key, v))
					}
					return
				},
			},
			// Computed
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {

	return nil
}

func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {

	return resourceDiskRead(d, m)
}

func resourceVMDelete(d *schema.ResourceData, m interface{}) error {

	return nil
}
