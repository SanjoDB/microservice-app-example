variable "resource_group_name" {
  type        = string
  description = "Nombre del Resource Group"
}

variable "location" {
  type        = string
  description = "Ubicación en Azure"
  default     = "eastus"
}

variable "admin_username" {
  type        = string
  description = "Nombre del usuario administrador"
}

variable "admin_password" {
  type        = string
  description = "Contraseña de la VM"
  sensitive   = true
}

variable "subscription_id" {
  type        = string
  description = "ID de la suscripción de Azure"
}