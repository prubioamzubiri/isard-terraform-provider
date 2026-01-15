# Ejemplo creando múltiples desktops

terraform {
  required_providers {
    isard = {
      source = "registry.terraform.io/tknika/isard"
    }
  }
}

provider "isard" {
  endpoint     = "localhost"
  auth_method  = "form"
  cathegory_id = "default"
  username     = "admin"
  password     = "IsardVDI"
}

# Obtener templates
data "isard_templates" "ubuntu" {
  name_filter = "Ubuntu"
}

# Crear múltiples desktops usando count
resource "isard_vm" "dev_team" {
  count = 3
  
  name        = "desktop-dev-${count.index + 1}"
  description = "Desktop para desarrollador ${count.index + 1}"
  template_id = data.isard_templates.ubuntu.templates[0].id
}

# Crear desktops usando for_each con nombres personalizados
locals {
  usuarios = {
    "juan"  = "Desktop de Juan"
    "maria" = "Desktop de María"
    "pedro" = "Desktop de Pedro"
  }
}

resource "isard_vm" "usuarios" {
  for_each = local.usuarios
  
  name        = "desktop-${each.key}"
  description = each.value
  template_id = data.isard_templates.ubuntu.templates[0].id
}

# Outputs
output "desktops_dev_team" {
  description = "IDs de los desktops del equipo dev"
  value = {
    for idx, desktop in isard_vm.dev_team :
    "dev-${idx + 1}" => {
      id     = desktop.id
      name   = desktop.name
      vcpus  = desktop.vcpus
      memory = desktop.memory
    }
  }
}

output "desktops_usuarios" {
  description = "IDs de los desktops de usuarios"
  value = {
    for user, desktop in isard_vm.usuarios :
    user => {
      id     = desktop.id
      name   = desktop.name
      vcpus  = desktop.vcpus
      memory = desktop.memory
    }
  }
}
