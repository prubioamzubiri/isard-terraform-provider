# Terraform Provider para Isard VDI

Provider de Terraform para gestionar recursos en Isard VDI a través de su API v3.

## Características

### Recursos

- ✅ **isard_vm** - Creación, lectura y eliminación de desktops persistentes con soporte para:
  - Hardware personalizado (vCPUs, memoria)
  - Interfaces de red personalizadas
- ✅ **isard_deployment** - Gestión de deployments para crear múltiples desktops para usuarios/grupos
- ✅ **isard_network** - Gestión de redes virtuales de usuario
- ✅ **isard_network_interface** - Gestión de interfaces de red del sistema (requiere admin)
- ✅ **isard_qos_net** - Gestión de perfiles QoS de red (requiere admin)

### Data Sources

- ✅ **isard_templates** - Listado de templates disponibles con filtrado por nombre
- ✅ **isard_network_interfaces** - Consulta de interfaces de red del sistema con filtros avanzados
- ✅ **isard_groups** - Consulta de grupos del sistema con filtrado por nombre y categoría

### Autenticación

- ✅ Soporte para autenticación mediante token JWT
- ✅ Soporte para autenticación mediante formulario (usuario/contraseña)
- ✅ Configuración SSL flexible para desarrollo y producción

## Requisitos

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (para desarrollo)
- Instancia de Isard VDI con API v3 activa

## Instalación

### Desarrollo Local

1. Clonar el repositorio:
```bash
git clone https://github.com/prubioamzubiri/isard-terraform-provider.git
cd isard-terraform-provider
```

2. Compilar el provider:
```bash
go build -o terraform-provider-isard
```

3. Configurar el override local en `~/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "tknika/isard" = "/home/tu-usuario/source/isard-terraform-provider"
  }
  direct {}
}
```

### Instalación desde Registry (futuro)

```hcl
terraform {
  required_providers {
    isard = {
      source  = "registry.terraform.io/tknika/isard"
      version = "~> 1.0"
    }
  }
}
```

## Uso Básico

```hcl
# Configurar el provider
provider "isard" {
  endpoint     = "localhost"
  auth_method  = "form"
  cathegory_id = "default"
  username     = "admin"
  password     = "IsardVDI"
}

# Obtener templates disponibles
data "isard_templates" "ubuntu" {
  name_filter = "Ubuntu"
}

# Obtener grupos por nombre
data "isard_groups" "desarrollo" {
  name_filter = "Desarrollo"
}

# Crear un desktop persistente
resource "isard_vm" "mi_desktop" {
  name        = "mi-desktop-terraform"
  description = "Desktop creado con Terraform"
  template_id = data.isard_templates.ubuntu.templates[0].id
}

# Crear un deployment para un equipo
resource "isard_deployment" "equipo_dev" {
  name         = "Deployment Equipo Dev"
  description  = "Desktops para el equipo de desarrollo"
  template_id  = data.isard_templates.ubuntu.templates[0].id
  desktop_name = "Desktop Dev"
  visible      = false
  
  vcpus  = 4
  memory = 8.0

  allowed {
    groups = [data.isard_groups.desarrollo.groups[0].id]
  }
}

# Crear una red virtual
resource "isard_network" "mi_red" {
  name        = "Red de Desarrollo"
  description = "Red virtual para desarrollo"
  model       = "virtio"
  qos_id      = "unlimited"
}

# Crear interfaz de red del sistema (requiere admin)
resource "isard_network_interface" "bridge_custom" {
  id          = "bridge-custom"
  name        = "Bridge Personalizado"
  description = "Bridge para entorno custom"
  net         = "br-custom"
  kind        = "bridge"
  model       = "virtio"
  qos_id      = "unlimited"
  
  # Hacer visible para todos los usuarios
  allowed {
    roles      = []
    categories = []
    groups     = []
    users      = []
  }
}

# Crear VM con interfaces personalizadas
resource "isard_vm" "vm_custom" {
  name        = "vm-con-red-custom"
  description = "VM con interfaces personalizadas"
  template_id = data.isard_templates.ubuntu.templates[0].id
  
  interfaces = [
    "wireguard",  # Requerido para RDP
    isard_network_interface.bridge_custom.id
  ]
}
```

## Documentación

### Provider

- [Configuración del Provider](docs/index.md)

### Recursos

- [Resource: isard_vm](docs/resources/isard_vm.md) - Gestión de VMs/desktops
- [Resource: isard_deployment](docs/resources/deployment.md) - Gestión de deployments
- [Resource: isard_network](docs/resources/isard_network.md) - Redes virtuales de usuario
- [Resource: isard_network_interface](docs/resources/isard_network_interface.md) - Interfaces de red del sistema
- [Resource: isard_qos_net](docs/resources/isard_qos_net.md) - Perfiles QoS de red

### Data Sources

- [Data Source: isard_templates](docs/data-sources/isard_templates.md) - Consulta de templates
- [Data Source: isard_network_interfaces](docs/data-sources/isard_network_interfaces.md) - Consulta de interfaces
- [Data Source: isard_groups](docs/data-sources/isard_groups.md) - Consulta de grupos

## Ejemplos

Consulta el directorio [examples/](examples/) para ver ejemplos completos de uso.

## Desarrollo

### Compilar

```bash
go build -o terraform-provider-isard
```

### Ejecutar Tests

```bash
go test ./...
```

### Depuración

Para habilitar logs detallados:

```bash
export TF_LOG=DEBUG
terraform plan
```

## Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/mi-feature`)
3. Commit tus cambios (`git commit -am 'Agregar nueva característica'`)
4. Push a la rama (`git push origin feature/mi-feature`)
5. Abre un Pull Request

## Licencia

[Especificar licencia]

## Soporte

Para reportar bugs o solicitar features, abre un issue en GitHub.
