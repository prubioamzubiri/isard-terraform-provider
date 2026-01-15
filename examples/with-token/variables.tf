variable "isard_token" {
  description = "Token JWT de Isard VDI para autenticación"
  type        = string
  sensitive   = true
  
  # No especificamos default por seguridad
  # El valor debe proporcionarse mediante:
  # - Variable de entorno: TF_VAR_isard_token
  # - Archivo terraform.tfvars
  # - Flag -var en la línea de comandos
}
