package gandi

import (
	"fmt"
	"regexp"

	"github.com/PabloPie/go-gandi/hosting"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSSH() *schema.Resource {
	return &schema.Resource{
		Create: resourceSSHCreate,
		Read:   resourceSSHRead,
		Delete: resourceSSHDelete,
		Exists: resourceSSHExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: sshKeyValidateName,
			},
			// Computed
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSSHCreate(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	sshkey, err := h.CreateKey(d.Get("name").(string), d.Get("value").(string))
	if err != nil {
		return err
	}
	d.SetId(sshkey.ID)
	d.Set("name", sshkey.Name)
	return resourceSSHRead(d, m)
}

func resourceSSHRead(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	sshkey := h.KeyFromName(d.Get("name").(string))
	if sshkey.ID == "" {
		d.SetId("")
		return nil
	}
	d.Set("name", sshkey.Name)
	d.Set("value", sshkey.Value)
	d.Set("fingerprint", sshkey.Fingerprint)
	return nil
}

func resourceSSHDelete(d *schema.ResourceData, m interface{}) error {
	h := m.(hosting.Hosting)
	sshkey := hosting.SSHKey{ID: d.Id()}
	err := h.DeleteKey(sshkey)
	return err
}

func resourceSSHExists(d *schema.ResourceData, m interface{}) (bool, error) {
	h := m.(hosting.Hosting)

	sshkey := h.KeyFromName(d.Get("name").(string))
	return sshkey.ID != "", nil
}

func sshKeyValidateName(value interface{}, name string) (warnings []string, errors []error) {
	r := regexp.MustCompile(`^(?:ssh-(?:rsa|dss|ed25519)|ecdsa-\S+) [A-Za-z0-9/+=]+(?: (\S+))?$`)
	if !r.Match([]byte(value.(string))) {
		errors = append(errors, fmt.Errorf("Invalid value: '%s', does not match %s", value.(string), r))
	}
	return
}
