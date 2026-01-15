# Ejemplo básico de uso del provider Isard

terraform {
  required_providers {
    isard = {
      source = "registry.terraform.io/tknika/isard"
    }
  }
}

# Configuración del provider
provider "isard" {
  endpoint     = "localhost"
  auth_method  = "form"
  cathegory_id = "default"
  username     = "admin"
  password     = "IsardVDI"
}

# Obtener templates disponibles
data "isard_templates" "all" {}

# Crear un desktop simple
resource "isard_vm" "ejemplo" {
  name        = "desktop-ejemplo"
  description = "Desktop de ejemplo creado con Terraform"
  template_id = data.isard_templates.all.templates[0].id
}

# Outputs
output "desktop_id" {
  description = "ID del desktop creado"
  value       = isard_vm.ejemplo.id
}

output "desktop_info" {
  description = "Información del desktop"
  value = {
    id          = isard_vm.ejemplo.id
    name        = isard_vm.ejemplo.name
    vcpus       = isard_vm.ejemplo.vcpus
    memory      = isard_vm.ejemplo.memory
    template_id = isard_vm.ejemplo.template_id
  }
}

output "templates_disponibles" {
  description = "Lista de todos los templates disponibles"
  value       = data.isard_templates.all.templates
}
