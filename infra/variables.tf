variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "devops-workshop-1"
}

variable "admin_username" {
  description = "Username for the VM admin user"
  type        = string
  default     = "azureuser"
}

variable "admin_password" {
  description = "Password for the admin user"
  type        = string
  sensitive   = true
}

variable "vm_size" {
  description = "Size of the virtual machine"
  type        = string
  default     = "Standard_DS1_v2"
}

variable "vnet_name" {
  description = "Name of the virtual network"
  type        = string
  default     = "vnet-devops"
}

variable "vnet_address_space" {
  description = "Address space for the virtual network"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "subnet_name" {
  description = "Name of the subnet"
  type        = string
  default     = "subnet-devops"
}

variable "subnet_address_prefixes" {
  description = "Address prefixes for the subnet"
  type        = list(string)
  default     = ["10.0.1.0/24"]
}

variable "public_ip_name" {
  description = "Name of the public IP"
  type        = string
  default     = "publicip-devops"
}

variable "public_ip_allocation_method" {
  description = "Allocation method for the public IP"
  type        = string
  default     = "Static"
}

variable "public_ip_sku" {
  description = "SKU for the public IP"
  type        = string
  default     = "Standard"
}

variable "nic_name" {
  description = "Name of the network interface"
  type        = string
  default     = "nic-devops"
}

variable "vm_name" {
  description = "Name of the virtual machine"
  type        = string
  default     = "vm-devops"
}

variable "storage_account_type" {
  description = "Storage account type for the OS disk"
  type        = string
  default     = "Standard_LRS"
}

variable "image_publisher" {
  description = "Publisher of the VM image"
  type        = string
  default     = "Canonical"
}

variable "image_offer" {
  description = "Offer of the VM image"
  type        = string
  default     = "UbuntuServer"
}

variable "image_sku" {
  description = "SKU of the VM image"
  type        = string
  default     = "18.04-LTS"
}

variable "image_version" {
  description = "Version of the VM image"
  type        = string
  default     = "latest"
}

variable "nsg_name" {
  description = "Name of the network security group"
  type        = string
  default     = "nsg-devops"
}