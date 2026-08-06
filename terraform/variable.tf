variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
  default     = "project-a76346dd-2461-4d0f-b24"
}

variable "region" {
  description = "Google Cloud region"
  type        = string
  default     = "europe-west1"
}

variable "mongodb_uri" {
  description = "MongoDB Atlas connection string for tour and payment services"
  type        = string
  sensitive   = true
}

variable "mongodb_uri_blog" {
  description = "MongoDB Atlas connection string for blog service"
  type        = string
  sensitive   = true
}

variable "neo4j_uri" {
  description = "Neo4j Aura connection URI"
  type        = string
  default     = "neo4j+s://25ac697b.databases.neo4j.io"
}

variable "neo4j_user" {
  description = "Neo4j username"
  type        = string
  sensitive   = true
}

variable "neo4j_password" {
  description = "Neo4j password"
  type        = string
  sensitive   = true
}

variable "postgres_password" {
  description = "Cloud SQL PostgreSQL password for stakeholders_user"
  type        = string
  sensitive   = true
}
