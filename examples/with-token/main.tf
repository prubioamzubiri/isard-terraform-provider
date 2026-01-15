# Ejemplo usando autenticación por token

terraform {
  required_providers {
    isard = {
      source = "registry.terraform.io/tknika/isard"
    }
  }
}

# Configuración usando token JWT
provider "isard" {
  endpoint     = "localhost"
  auth_method  = "token"
  cathegory_id = "default"
  token        = var.isard_token
}

# Obtener templates
data "isard_templates" "all" {}

# Crear desktop
resource "isard_vm" "token_auth_example" {
  name        = "desktop-token-auth"
  description = "Desktop creado usando autenticación por token"
  template_id = data.isard_templates.all.templates[0].id
}

# Outputs
output "desktop_creado" {
  value = {
    id          = isard_vm.token_auth_example.id
    name        = isard_vm.token_auth_example.name
    description = isard_vm.token_auth_example.description
  }
}
