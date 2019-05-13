package gandi

import (
	"errors"
	"fmt"
	"log"

	"github.com/PabloPie/Gandi-Go/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Exists: resourceVMExists,

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
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Names of the ssh keys allowed to connect",
			},
			"userpass": {
				Type:     schema.TypeList,
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the disk to use as boot disk, must exist and be free",
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
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
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
				Required: true,
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	var vm hosting.VM

	vmspec, err := parseVMSpec(d, h)
	if err != nil {
		return err
	}
	bootdisk, err := parseBootDisk(d, h)
	if err != nil {
		return err
	}
	ips, err := parseIPS(d, h)
	if err != nil {
		return err
	}
	vm, _, bootdisk, err = h.CreateVMWithExistingDiskAndIP(vmspec, ips[0], bootdisk)
	if err != nil {
		return err
	}

	// Attach remaining ips, the first one was attached on creation
	for _, ip := range ips[1:] {
		log.Printf("[INFO] Attaching ip '%s' to vm '%s'...", ip.IP, vm.Hostname)
		vm, ip, err := h.AttachIP(vm, ip)
		if err != nil {
			log.Printf("[WARN] Error attaching ip '%s' to vm '%s': %s", ip.ID, vm.Hostname, err)
		}
	}

	disks := parseDisks(d, h)
	for _, disk := range disks {
		log.Printf("[INFO] Attaching disk '%s' to vm '%s'...", disk.Name, vm.Hostname)
		vm, disk, err := h.AttachDisk(vm, disk)
		if err != nil {
			log.Printf("[WARN] Error attaching disk '%s' to vm '%s': %s", disk.Name, vm.Hostname, err)
		}
	}

	d.SetId(vm.ID)
	d.Set("name", vm.Hostname)
	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vm, err := h.VMFromName(d.Get("name").(string))
	// XXX: add retry option
	if err != nil {
		return err
	}

	d.Set("memory", vm.Memory)
	if vm.Farm != "" {
		d.Set("farm", vm.Farm)
	}
	d.Set("cores", vm.Cores)
	d.Set("state", vm.State)
	var ips []map[string]interface{}
	for _, ip := range vm.Ips {
		ips = append(
			ips,
			map[string]interface{}{
				"id": ip.ID,
				"ip": ip.IP,
			},
		)
	}
	var disks []map[string]interface{}
	for _, disk := range vm.Disks {
		disks = append(
			disks,
			map[string]interface{}{
				"id":   disk.ID,
				"name": disk.Name,
				"size": disk.Size,
			},
		)
	}
	d.Set("ips", ips)
	d.Set("disks", disks)
	return nil
}

// Need to be considered for update:
// easy: memory, cores, state, name
// harder: disks, ip, boot_disk_id
func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {

	return resourceDiskRead(d, m)
}

// Deleting a vm deletes its boot disk and IP
func resourceVMDelete(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vm := hosting.VM{ID: d.Id()}
	if err := h.StopVM(vm); err != nil {
		return err
	}
	if err := h.DeleteVM(vm); err != nil {
		return err
	}
	return nil
}

func resourceVMExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)

	_, err := h.VMFromName(d.Get("name").(string))
	if err != nil {
		return false, err
	}
	return true, nil
}

func parseVMSpec(d *schema.ResourceData, h hosting.Hosting) (vmspec hosting.VMSpec, err error) {
	vmspec.RegionID = d.Get("region_id").(string)
	if name, ok := d.GetOk("name"); ok {
		vmspec.Hostname = name.(string)
	}
	if farm, ok := d.GetOk("farm"); ok {
		vmspec.Farm = farm.(string)
	}
	if memory, ok := d.GetOk("memory"); ok {
		vmspec.Memory = memory.(int)
	}
	if cores, ok := d.GetOk("cores"); ok {
		vmspec.Cores = cores.(int)
	}
	if sshkeys, ok := d.GetOk("ssh_keys"); ok {
		for _, sshkey := range sshkeys.([]string) {
			key := h.KeyFromName(sshkey)
			if key.ID == "" {
				log.Printf("[WARN] Key '%s' not found", sshkey)
				continue
			}
			vmspec.SSHKeysID = append(vmspec.SSHKeysID, key.ID)
		}
	}
	if rawuserpass, ok := d.GetOk("userpass"); ok {
		userpasslist := rawuserpass.([]interface{})
		userpass := userpasslist[0].(map[string]interface{})
		vmspec.Login = userpass["login"].(string)
		vmspec.Password = userpass["password"].(string)
	}
	if len(vmspec.SSHKeysID) < 1 && vmspec.Login == "" {
		err = errors.New("SSH keys or login/password required but not provided")
	}
	return
}

func parseBootDisk(d *schema.ResourceData, h hosting.Hosting) (bootdisk hosting.Disk, err error) {
	bootdiskid := d.Get("boot_disk_id").(string)
	disks, err := h.DescribeDisks(hosting.DiskFilter{ID: bootdiskid})
	if err != nil {
		return hosting.Disk{}, err
	}
	if len(disks) < 1 {
		err = fmt.Errorf("Error, Disk '%s' not found", bootdiskid)
	}
	bootdisk = disks[0]
	return
}

func parseIPS(d *schema.ResourceData, h hosting.Hosting) (ips []hosting.IPAddress, err error) {
	ipslist, ok := d.GetOk("ips")
	if !ok {
		return nil, errors.New("Error, no ips set but a minimum of 1 is required")
	}

	for _, rawip := range ipslist.([]interface{}) {
		ipmap := rawip.(map[string]interface{})
		ip, err := h.DescribeIP(hosting.IPFilter{ID: ipmap["id"].(string)})
		if err != nil {
			return nil, err
		}
		if len(ip) < 1 {
			log.Printf("[ERR] IP with id %s not found", ipmap["id"])
			continue
		}
		ips = append(ips, ip[0])
	}
	// ips were provided but none are actually valid
	if len(ips) < 1 {
		return nil, errors.New("Error, ips provided, but none were found")
	}
	return
}

func parseDisks(d *schema.ResourceData, h hosting.Hosting) (disks []hosting.Disk) {
	rawdisks := d.Get("disks")

	for _, rawdisk := range rawdisks.([]interface{}) {
		diskmap := rawdisk.(map[string]interface{})
		disk := h.DiskFromName(diskmap["name"].(string))
		if disk.ID != "" {
			disks = append(disks, disk)
		}
	}
	return
}
