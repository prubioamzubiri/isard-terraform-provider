variable "isard_endpoint" {
  description = "Hostname o IP del servidor Isard VDI"
  type        = string
  default     = "localhost"
}

variable "isard_auth_method" {
  description = "Método de autenticación (form o token)"
  type        = string
  default     = "form"
  
  validation {
    condition     = contains(["form", "token"], var.isard_auth_method)
    error_message = "El método de autenticación debe ser 'form' o 'token'."
  }
}

variable "isard_category" {
  description = "ID de la categoría en Isard VDI"
  type        = string
  default     = "default"
}

variable "isard_username" {
  description = "Nombre de usuario para autenticación (requerido si auth_method=form)"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "isard_password" {
  description = "Contraseña para autenticación (requerido si auth_method=form)"
  type        = string
  default     = "IsardVDI"
  sensitive   = true
}

variable "isard_token" {
  description = "Token JWT para autenticación (requerido si auth_method=token)"
  type        = string
  default     = ""
  sensitive   = true
}
