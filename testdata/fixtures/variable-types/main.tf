# Fixture: variable-types
# Exercises structured variable types (TypeSpec) and decoded constant
# defaults (DefaultValue): every primitive, collection, structural, and
# `any` type kind, plus optional object attributes with and without
# defaults, and a spread of literal default values including a
# precision-sensitive number and an explicit null.

# --- structured types (item: structured variable types) ---

variable "name" {
  type = string
}

variable "count" {
  type = number
}

variable "enabled" {
  type = bool
}

variable "tags" {
  type = map(string)
}

variable "subnets" {
  type = list(string)
}

variable "ids" {
  type = set(string)
}

variable "pair" {
  type = tuple([string, number])
}

variable "any_val" {
  type = any
}

variable "config" {
  type = object({
    name   = string
    size   = optional(number, 10)
    nested = optional(object({ a = string }))
  })
}

# --- decoded constant defaults (item: typed default values) ---

variable "default_string" {
  default = "hello"
}

variable "default_number" {
  default = 42
}

variable "default_big_number" {
  # Precision check: must survive as an exact decimal string.
  default = 1234567890123456789
}

variable "default_bool" {
  default = true
}

variable "default_list" {
  default = ["a", "b"]
}

variable "default_map" {
  default = {
    env  = "prod"
    tier = "1"
  }
}

variable "default_object" {
  default = {
    name = "web"
    port = 8080
  }
}

variable "default_null" {
  type    = string
  default = null
}

variable "no_default" {
  type = string
}
