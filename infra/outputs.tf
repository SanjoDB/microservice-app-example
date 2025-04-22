output "vm_ip_address" {
  description = "Dirección IP privada de la VM"
  value       = azurerm_network_interface.nic.private_ip_address
}

output "vm_public_ip" {
  description = "Dirección IP pública de la VM"
  value       = azurerm_public_ip.public_ip.ip_address
}