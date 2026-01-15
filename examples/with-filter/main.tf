# Ejemplo usando filtros de templates

terraform {
  required_providers {
    isard = {
      source = "registry.terraform.io/tknika/isard"
    }
  }
}

provider "isard" {
  endpoint     = var.isard_endpoint
  auth_method  = var.isard_auth_method
  cathegory_id = var.isard_category
  username     = var.isard_username
  password     = var.isard_password
}

# Filtrar templates Ubuntu
data "isard_templates" "ubuntu" {
  name_filter = "Ubuntu"
}

# Filtrar templates Windows
data "isard_templates" "windows" {
  name_filter = "Windows"
}

# Crear desktop Ubuntu
resource "isard_vm" "ubuntu_desktop" {
  count = length(data.isard_templates.ubuntu.templates) > 0 ? 1 : 0
  
  name        = "desktop-ubuntu-${count.index + 1}"
  description = "Desktop Ubuntu para desarrollo"
  template_id = data.isard_templates.ubuntu.templates[0].id
}

# Crear desktop Windows
resource "isard_vm" "windows_desktop" {
  count = length(data.isard_templates.windows.templates) > 0 ? 1 : 0
  
  name        = "desktop-windows-${count.index + 1}"
  description = "Desktop Windows para testing"
  template_id = data.isard_templates.windows.templates[0].id
}

# Outputs
output "ubuntu_templates" {
  description = "Templates Ubuntu encontrados"
  value = [
    for t in data.isard_templates.ubuntu.templates : {
      id   = t.id
      name = t.name
    }
  ]
}

output "windows_templates" {
  description = "Templates Windows encontrados"
  value = [
    for t in data.isard_templates.windows.templates : {
      id   = t.id
      name = t.name
    }
  ]
}

output "desktops_creados" {
  description = "IDs de los desktops creados"
  value = {
    ubuntu  = length(isard_vm.ubuntu_desktop) > 0 ? isard_vm.ubuntu_desktop[0].id : null
    windows = length(isard_vm.windows_desktop) > 0 ? isard_vm.windows_desktop[0].id : null
  }
}
