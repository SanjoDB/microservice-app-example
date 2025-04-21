output "vm_ip_address" {
  description = "Dirección IP privada de la VM"
  value       = azurerm_network_interface.nic.private_ip_address
}