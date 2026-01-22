# Data Source: isard_groups

Obtiene la lista de grupos disponibles en Isard VDI. Este data source es útil para encontrar grupos por nombre y obtener sus IDs para usarlos en recursos como `isard_deployment`.

## Ejemplo de Uso

### Obtener Todos los Grupos

```hcl
data "isard_groups" "todos" {}

output "lista_completa" {
  value = data.isard_groups.todos.groups
}
```

### Filtrar Grupos por Nombre

```hcl
data "isard_groups" "wag_p" {
  name_filter = "WAG-P"
}

output "grupo_wag_p" {
  value = data.isard_groups.wag_p.groups
}
```

### Usar con Deployment Resource

```hcl
data "isard_groups" "desarrollo" {
  name_filter = "Desarrollo"
}

data "isard_templates" "ubuntu" {
  name_filter = "Ubuntu Desktop"
}

resource "isard_deployment" "deploy_desarrollo" {
  name         = "Deployment para Desarrollo"
  description  = "Desktops para el equipo de desarrollo"
  template_id  = data.isard_templates.ubuntu.templates[0].id
  desktop_name = "Desktop-Dev"
  
  allowed {
    groups = [data.isard_groups.desarrollo.groups[0].id]
  }
}
```

### Filtrado Case-Insensitive

```hcl
# Estos tres ejemplos devolverán los mismos resultados
data "isard_groups" "test1" {
  name_filter = "wag-p"
}

data "isard_groups" "test2" {
  name_filter = "WAG-P"
}

data "isard_groups" "test3" {
  name_filter = "WaG-p"
}
```

### Filtrar por Categoría

```hcl
data "isard_groups" "grupos_default" {
  category_id = "default"
}

output "grupos_categoria_default" {
  value = data.isard_groups.grupos_default.groups
}
```

### Combinar Filtros

```hcl
# Buscar grupos que contengan "Dev" en la categoría "default"
data "isard_groups" "dev_default" {
  name_filter = "Dev"
  category_id = "default"
}

output "grupos_dev" {
  value = data.isard_groups.dev_default.groups
}
```

### Verificar que Exista al Menos un Grupo

```hcl
data "isard_groups" "produccion" {
  name_filter = "Produccion"
}

# Fallar si no hay grupos
locals {
  group_count = length(data.isard_groups.produccion.groups)
  group_id    = local.group_count > 0 ? data.isard_groups.produccion.groups[0].id : null
}

resource "isard_deployment" "prod_deploy" {
  count = local.group_count > 0 ? 1 : 0
  
  name         = "Deployment Producción"
  description  = "Desktops para producción"
  template_id  = "template-id"
  desktop_name = "Desktop-Prod"
  
  allowed {
    groups = [local.group_id]
  }
}
```

## Argumentos

Los siguientes argumentos son soportados:

### Opcionales

- `name_filter` - (Opcional) Filtro para buscar grupos por nombre. La búsqueda es case-insensitive y busca coincidencias parciales (substring). Si no se especifica junto con category_id, devuelve todos los grupos disponibles.

- `category_id` - (Opcional) Filtro para buscar grupos por categoría. Debe ser el ID exacto de la categoría. Se puede combinar con name_filter.

## Atributos Exportados

- `id` - ID del data source (siempre es `"groups"`).
- `groups` - Lista de objetos group. Cada objeto group contiene:
  - `id` - ID único del grupo.
  - `name` - Nombre del grupo.
  - `description` - Descripción del grupo.
  - `parent_category` - ID de la categoría padre a la que pertenece el grupo.

## Comportamiento del Filtrado

El filtrado funciona de la siguiente manera:

1. **Sin filtros:** Devuelve todos los grupos disponibles en el sistema
2. **Con name_filter:** Devuelve solo los grupos cuyo nombre contenga el string especificado
3. **Con category_id:** Devuelve solo los grupos de la categoría especificada
4. **Con ambos filtros:** Devuelve grupos que cumplan ambas condiciones
5. **Case-insensitive:** El filtrado por nombre no distingue entre mayúsculas y minúsculas
6. **Substring match:** Busca el filtro como parte del nombre, no requiere coincidencia exacta

### Ejemplos de Filtrado

Si tienes estos grupos:
- "Default" (categoría: default)
- "WAG-P" (categoría: default)
- "Desarrolladores" (categoría: tecnologia)
- "Soporte IT" (categoría: tecnologia)

```hcl
# Devuelve: WAG-P
data "isard_groups" "wag" {
  name_filter = "WAG"
}

# Devuelve: Desarrolladores, Soporte IT
data "isard_groups" "tech" {
  category_id = "tecnologia"
}

# Devuelve: Default, WAG-P
data "isard_groups" "default_cat" {
  category_id = "default"
}

# Devuelve: Soporte IT
data "isard_groups" "soporte" {
  name_filter = "soporte"
  category_id = "tecnologia"
}

# Devuelve: todos los grupos
data "isard_groups" "todos" {}
```

## Ejemplos Adicionales

### Seleccionar Grupo Específico por Nombre Exacto

```hcl
data "isard_groups" "busqueda" {
  name_filter = "WAG"
}

locals {
  # Buscar un grupo específico por nombre exacto
  wag_p_group = [
    for g in data.isard_groups.busqueda.groups : g
    if g.name == "WAG-P"
  ][0]
}

resource "isard_deployment" "deploy" {
  name         = "Deployment para WAG-P"
  template_id  = "template-id"
  desktop_name = "Desktop-WAG"
  
  allowed {
    groups = [local.wag_p_group.id]
  }
}
```

### Listar Información de Grupos

```hcl
data "isard_groups" "todos" {}

output "grupos_info" {
  value = {
    for group in data.isard_groups.todos.groups :
    group.name => {
      id          = group.id
      description = group.description
      category    = group.parent_category
    }
  }
}
```

### Crear Deployment para Múltiples Grupos

```hcl
data "isard_groups" "equipos" {
  name_filter = "Equipo"
}

data "isard_templates" "ubuntu" {
  name_filter = "Ubuntu"
}

resource "isard_deployment" "multi_equipo" {
  for_each = { 
    for idx, grp in data.isard_groups.equipos.groups : 
    grp.name => grp 
  }
  
  name         = "Deployment-${each.key}"
  description  = "Desktops para ${each.key}"
  template_id  = data.isard_templates.ubuntu.templates[0].id
  desktop_name = "Desktop-${each.key}"
  
  allowed {
    groups = [each.value.id]
  }
}
```

### Obtener IDs de Grupos para Permisos

```hcl
data "isard_groups" "admin" {
  name_filter = "Administradores"
}

data "isard_groups" "users" {
  name_filter = "Usuarios"
}

resource "isard_network_interface" "shared_network" {
  name        = "Red Compartida"
  description = "Red accesible por admins y usuarios"
  net         = "br-shared"
  kind        = "bridge"
  model       = "virtio"
  qos_id      = "unlimited"
  
  allowed {
    groups = concat(
      [for g in data.isard_groups.admin.groups : g.id],
      [for g in data.isard_groups.users.groups : g.id]
    )
  }
}
```

## Notas

- Este data source requiere permisos de administrador para acceder al endpoint `/api/v3/admin/groups`.
- Los grupos devueltos incluyen información sobre grupos enlazados (linked groups).
- Si un grupo no tiene descripción, el campo `description` contendrá un string vacío o un valor por defecto como `"[Default] "`.
- El filtrado se aplica en el lado del proveedor después de obtener todos los grupos de la API, no en el servidor.
