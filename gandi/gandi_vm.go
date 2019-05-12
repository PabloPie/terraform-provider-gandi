package gandi

import (
	"fmt"

	"github.com/PabloPie/Gandi-Go/hosting"
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
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Names of the ssh keys allowed to connect",
			},
			"userpass": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
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
			// delete? this resource is not managed by terraform
			"boot_disk": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"boot_disk_id"},
				Description:   "Disk spec of the disk used as boot disk",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"disk_image_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// "delete_on_detach"
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
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ip_version"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ip_version": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"ips"},
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v != 4 && v != 6 {
						errs = append(errs, fmt.Errorf("%q must be either 4 or 6, got: %d", key, v))
					}
					return
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	// h := m.(hosting.Hosting)
	// vmspec, err := parseVMSpec(d)
	// diskspec, err := parseDiskSpec(d)

	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

// Need to be considered for update:
// easy: memory, cores, state, name
// harder: disks, ip, boot_disk, boot_disk_id
func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {

	return resourceDiskRead(d, m)
}

func resourceVMDelete(d *schema.ResourceData, m interface{}) error {

	return nil
}

func parseVMSpec(d *schema.ResourceData) (vmspec hosting.VMSpec, err error) {
	return
}
