package gandi

import (
	"errors"
	"fmt"
	"log"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Exists: resourceVMExists,
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
			// keys and login can change on boot, c.f vm.start()
			// Modification requires stopping and starting the machine
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
			"boot_disk": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
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
			"disks": {
				Type:     schema.TypeSet,
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
			"ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
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

	vmspec, err := parseVMSpec(d)
	if err != nil {
		return err
	}
	ipslist := d.Get("ips").(*schema.Set).List()
	ips, err := parseIPS(h, ipslist)
	if err != nil {
		return err
	}
	bootdiskraw := d.Get("boot_disk").([]interface{})
	bootdisk, err := parseDisks(h, bootdiskraw)
	if err != nil {
		return err
	}

	vm, _, _, err = h.CreateVMWithExistingDiskAndIP(vmspec, ips[0], bootdisk[0])
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

	disklist := d.Get("disks").(*schema.Set).List()
	disks, err := parseDisks(h, disklist)
	if err != nil {
		return err
	}

	// Attach non-boot disks, the final disk list will not contain the boot disk
	for _, disk := range disks {
		log.Printf("[INFO] Attaching disk '%s' to vm '%s'...", disk.Name, vm.Hostname)
		vm, disk, err := h.AttachDisk(vm, disk)
		if err != nil {
			log.Printf("[WARN] Error attaching disk '%s' to vm '%s': %s", disk.Name, vm.Hostname, err)
		}
	}

	d.SetId(vm.ID)
	d.Set("ssh_keys", vm.SSHKeysID)
	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vms, err := h.DescribeVM(hosting.VMFilter{ID: d.Id()})
	if err != nil {
		return err
	}
	// No vm with that ID exists
	if len(vms) < 1 {
		d.SetId("")
		return nil
	}
	vm := vms[0]
	d.Set("name", vm.Hostname)
	d.Set("region_id", vm.RegionID)
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
	// Disk at position 0 is the boot disk
	d.Set("disks", disks[1:])
	d.Set("boot_disk", disks[:1])
	return nil
}

func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vm := hosting.VM{ID: d.Id(), RegionID: d.Get("region_id").(string)}
	d.Partial(true)
	if d.HasChange("memory") {
		_, newmem := d.GetChange("memory")
		vmupdated, err := h.UpdateVMMemory(vm, newmem.(int))
		if err != nil {
			return fmt.Errorf("[ERR] Memory update failed: %s", err)
		}
		log.Printf("[INFO] Memory for vm '%s' updated to %dGB", vmupdated.Hostname, vmupdated.Memory)
		d.SetPartial("memory")
		vm = vmupdated
	}
	if d.HasChange("cores") {
		_, newcores := d.GetChange("cores")
		vmupdated, err := h.UpdateVMCores(vm, newcores.(int))
		if err != nil {
			return fmt.Errorf("[ERR] Updating number of cores failed: %s", err)
		}
		log.Printf("[INFO] Number of cores for vm '%s' updated to %d", vmupdated.Hostname, vmupdated.Cores)
		d.SetPartial("cores")
		vm = vmupdated
	}
	if d.HasChange("state") {
		_, newstate := d.GetChange("state")
		state := newstate.(string)
		var err error
		switch state {
		case "halted":
			err = h.StopVM(vm)
		case "running":
			err = h.StartVM(vm)
		case "deleted":
			err = resourceVMDelete(d, m)
		default:
			return fmt.Errorf("[WARN] Invalid option for state '%s'", state)
		}
		if err != nil {
			return fmt.Errorf("[ERR] Operation on VM failed: %s", err)
		}
		log.Printf("[INFO] Operation on VM successful")
		d.SetPartial("state")
		vm.State = state
	}
	if d.HasChange("name") {
		_, newname := d.GetChange("name")
		vmupdated, err := h.RenameVM(vm, newname.(string))
		if err != nil {
			return fmt.Errorf("[ERR] Error renaming VM '%s'", vm.Hostname)
		}
		log.Printf("[INFO] VM '%s' renamed to '%s'", vm.Hostname, vmupdated.Hostname)
		d.SetPartial("name")
		vm = vmupdated
	}
	if d.HasChange("boot_disk") {
		oldbootdisk, newbootdisk := d.GetChange("boot_disk")
		olddisk, err := parseDisks(h, oldbootdisk.([]interface{}))
		if err != nil {
			return err
		}
		newdisk, err := parseDisks(h, newbootdisk.([]interface{}))
		if err != nil {
			return err
		}
		// Attaching to position 0 still leaves the other disk attached
		vmupdated, _, err := h.AttachDiskAtPosition(vm, newdisk[0], 0)
		if err != nil {
			return err
		}
		vmupdated, _, err = h.DetachDisk(vmupdated, olddisk[0])
		if err != nil {
			return err
		}
		d.SetPartial("boot_disk")
	}
	if d.HasChange("disks") {
		olddisks, newdisks := d.GetChange("disks")
		olddisklist, err := parseDisks(h, olddisks.(*schema.Set).List())
		if err != nil {
			return err
		}
		newdisklist, err := parseDisks(h, newdisks.(*schema.Set).List())
		if err != nil {
			return err
		}
		todetach, toattach := diskDiff(olddisklist, newdisklist)
		for _, disk := range todetach {
			vmupdated, _, err := h.DetachDisk(vm, disk)
			if err != nil {
				return fmt.Errorf("[ERR] Could not detach Disk '%s': %s", disk.Name, err)
			}
			vm = vmupdated
		}
		for _, disk := range toattach {
			vmupdated, _, err := h.AttachDisk(vm, disk)
			if err != nil {
				return fmt.Errorf("[ERR] Could not attach Disk '%s': %s", disk.Name, err)
			}
			vm = vmupdated
		}
		d.SetPartial("disks")
	}
	if d.HasChange("ips") {
		oldips, newips := d.GetChange("ips")
		oldiplist, err := parseIPS(h, oldips.(*schema.Set).List())
		if err != nil {
			return err
		}
		newiplist, err := parseIPS(h, newips.(*schema.Set).List())
		if err != nil {
			return err
		}
		todetach, toattach := ipDiff(oldiplist, newiplist)
		for _, ip := range todetach {
			vmupdated, ip, err := h.DetachIP(vm, ip)
			if err != nil {
				return fmt.Errorf("[ERR] Could not detach IP '%s': %s", ip.IP, err)
			}
			vm = vmupdated
		}
		for _, ip := range toattach {
			vmupdated, _, err := h.AttachIP(vm, ip)
			if err != nil {
				return fmt.Errorf("[ERR] Could not attach IP '%s': %s", ip.IP, err)
			}
			vm = vmupdated
		}
		d.SetPartial("ips")
	}
	d.Partial(false)
	return resourceVMRead(d, m)
}

// Deleting a vm does not delete its boot disk nor any of its ips
func resourceVMDelete(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	vm := hosting.VM{ID: d.Id(), RegionID: d.Get("region_id").(string)}
	var err error
	if exists, _ := resourceVMExists(d, m); !exists {
		return nil
	}
	h.StopVM(vm)
	// detach ips and disks to avoid deletion
	bootdisk := d.Get("boot_disk").([]interface{})[0].(map[string]interface{})
	disk := h.DiskFromName(bootdisk["name"].(string))
	_, _, err = h.DetachDisk(vm, disk)
	if err != nil {
		return fmt.Errorf("[ERR] Could not detach disk '%s': %s", bootdisk["name"].(string), err)
	}
	iplist := d.Get("ips").(*schema.Set).List()
	for _, ipraw := range iplist {
		ipmap := ipraw.(map[string]interface{})
		ipid := ipmap["id"].(string)
		ip := hosting.IPAddress{ID: ipid, RegionID: d.Get("region_id").(string)}
		_, _, err = h.DetachIP(vm, ip)
		if err != nil {
			return fmt.Errorf("[ERR] Could not detach IP '%s'(%s): %s", ip.IP, ipid, err)
		}
	}
	if err := h.DeleteVM(vm); err != nil {
		return err
	}
	return nil
}

func resourceVMExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)
	vms, err := h.DescribeVM(hosting.VMFilter{ID: d.Id()})
	return (err == nil && len(vms) > 0), err
}

func parseVMSpec(d *schema.ResourceData) (vmspec hosting.VMSpec, err error) {
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
		rawkeys := sshkeys.(*schema.Set).List()
		for _, key := range rawkeys {
			vmspec.SSHKeysID = append(vmspec.SSHKeysID, key.(string))
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

func parseIPS(h hosting.Hosting, iplist []interface{}) (ips []hosting.IPAddress, err error) {
	for _, rawip := range iplist {
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
	// at least 1 ip was provided but none are actually valid
	if len(ips) < 1 {
		return nil, errors.New("Error, ips provided, but none were found")
	}
	return
}

func parseDisks(h hosting.Hosting, disklist []interface{}) (disks []hosting.Disk, err error) {
	for _, rawdisk := range disklist {
		diskmap := rawdisk.(map[string]interface{})
		disk := h.DiskFromName(diskmap["name"].(string))
		if disk.ID != "" {
			disks = append(disks, disk)
		}
	}
	// at least 1 disk was provided but none are actually valid
	if len(disks) < 1 {
		return nil, errors.New("Error, disks provided, but none were found")
	}
	return
}

// needs testing
func ipDiff(oldips []hosting.IPAddress, newips []hosting.IPAddress) (todetach []hosting.IPAddress, toattach []hosting.IPAddress) {
	for _, oldip := range oldips {
		found := false
		for i, newip := range newips {
			if oldip.ID == newip.ID {
				found = true
				// no attach or detach operation for an ip that was already attached
				// delete this ip from the list
				newips = append(newips[:i], newips[i+1:]...)
				break
			}
		}
		// an ip previously attached (found in oldips) is no longer present
		// in newips, add to the list to be detached
		if !found {
			todetach = append(todetach, oldip)
		}
	}
	// disks that need to be attached are the disks that are on the new list of ips
	// but not the old one
	toattach = newips
	return
}

func diskDiff(olddisks []hosting.Disk, newdisks []hosting.Disk) (todetach []hosting.Disk, toattach []hosting.Disk) {
	for _, olddisk := range olddisks {
		found := false
		for i, newdisk := range newdisks {
			if olddisk.ID == newdisk.ID {
				found = true
				newdisks = append(newdisks[:i], newdisks[i+1:]...)
				break
			}
		}
		if !found {
			todetach = append(todetach, olddisk)
		}
	}
	toattach = newdisks
	return
}
