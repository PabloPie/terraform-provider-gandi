package gandi

import (
	c "github.com/PabloPie/Gandi-Go/client"
	"github.com/PabloPie/Gandi-Go/hosting/hostingv4"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	// Hosting requires an API key
	// Default URL is for v4, a different URL can also be used
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GANDI_API_KEY", nil),
				Description: "Gandi API Key",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GANDI_API_URL", ""),
				Description: "Gandi API URL to use for requests",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"gandi_region": dataSourceRegion(),
			"gandi_image":  dataSourceImage(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"gandi_disk": resourceDisk(),
			"gandi_ip":   resourceIP(),
		},
		ConfigureFunc: getGandiClient,
	}
}

// we need to make this variable for hosting, livedns, domain...
func getGandiClient(d *schema.ResourceData) (interface{}, error) {
	gandiClient, _ := c.NewClientv4(d.Get("url").(string), d.Get("api_key").(string))
	gandiHosting := hostingv4.Newv4Hosting(gandiClient)
	return gandiHosting, nil
}
