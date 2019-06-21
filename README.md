# Gandi Hosting Terraform Provider

This Terraform provider can be used to manage resources on Gandi's Hosting service. It currently supports Disks (from OS Image and data), IPAddresses (public and private), Vlans, SSH Keys and Virtual machines. Data sources for Disk Images and Regions (Datacenters) are also implemented.

## Usage example

This code sample contains an example for each type of resource
```
## Setting the env variable GANDI_API_KEY also works
provider "gandi" {
  api_key = "YOUR-API-KEY"
}

## DATA SOURCES

# REGION
data "gandi_region" "datacenter" {
    region_code = "FR-SD6"
}

# DISK IMAGE
data "gandi_image" "debian9" {
  name = "Debian 9"
  region_id = "${data.gandi_region.datacenter.id}"
}

## RESOURCES

# VLAN
resource "gandi_vlan" "vlan1" {
  name = "newnamevlan"
  region_id = "${data.gandi_region.datacenter.id}"
  subnet = "192.168.1.0/24"
  gateway = "192.168.1.1"
}

# PRIVATE IP
resource "gandi_private_ip" "ip1" {
  region_id = "${data.gandi_region.datacenter.id}"
  vlan_id = "${gandi_vlan.vlan1.id}"
  ip = "192.168.1.2"
}

# PUBLIC IP
resource "gandi_ip" "ip1" {
  region_id = "${data.gandi_region.datacenter.id}"
  version = 4
}

# BOOT DISK
resource "gandi_disk" "sd1" {
  region_id = "${data.gandi_region.datacenter.id}"
  src_disk_id = "${data.gandi_image.debian9.disk_id}"
  size = 10
  name = "d9_sysdisk"
}

# DATA DISK
resource "gandi_disk" "data1" {
  region_id = "${data.gandi_region.datacenter.id}"
  size = 20
  name = "datadisk"
}

# SSH KEY
resource "gandi_ssh" "sshkey1" {
  name = "mysshkey1"
  value = "ssh-rsa AAAA test@test"
}

# VIRTUAL MACHINE
resource "gandi_vm" "vm1" {
  region_id = "${data.gandi_region.datacenter.id}"
  name = "VM1"
  ips {
    id = "${gandi_ip.ip1.id}"
  }

  ips {
    id = "${gandi_private_ip.ip1.id}"
  }

  boot_disk {
    name = "${gandi_disk.sd1.name}"
  }

  disks {
    name = "${gandi_disk.data1.name}"
  }

  ssh_keys = ["${gandi_ssh.sshkey1.name}"]
}
```
