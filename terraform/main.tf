terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "run" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "sqladmin" {
  service            = "sqladmin.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "cloudbuild" {
  service            = "cloudbuild.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "containerregistry" {
  service            = "containerregistry.googleapis.com"
  disable_on_destroy = false
}


resource "google_sql_database_instance" "postgres" {
  name             = "soa-tourism-postgres-eu"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled = true

      authorized_networks {
        name  = "cloud-run"
        value = "0.0.0.0/0"
      }
    }
  }

  deletion_protection = false
  depends_on          = [google_project_service.sqladmin]
}

resource "google_sql_database" "stakeholders_db" {
  name     = "stakeholders_db"
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "stakeholders_user" {
  name     = "stakeholders_user"
  instance = google_sql_database_instance.postgres.name
  password = var.postgres_password
}


# API Gateway
resource "google_cloud_run_v2_service" "api_gateway" {
  name     = "api-gateway"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/api-gateway:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "TOUR_SERVICE_URL"
        value = google_cloud_run_v2_service.tour_service.uri
      }
      env {
        name  = "FOLLOWER_SERVICE_URL"
        value = google_cloud_run_v2_service.follower_service.uri
      }
      env {
        name  = "PAYMENT_SERVICE_URL"
        value = google_cloud_run_v2_service.payment_service.uri
      }
      env {
        name  = "BLOG_SERVICE_URL"
        value = google_cloud_run_v2_service.blog_service.uri
      }
      env {
        name  = "STAKEHOLDERS_SERVICE_URL"
        value = google_cloud_run_v2_service.stakeholders_service.uri
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.run]
}

# Tour Service
resource "google_cloud_run_v2_service" "tour_service" {
  name     = "tour-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/tour-service:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "MONGODB_URI"
        value = var.mongodb_uri
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.run]
}

# Follower Service
resource "google_cloud_run_v2_service" "follower_service" {
  name     = "follower-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/follower-service:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "NEO4J_URI"
        value = var.neo4j_uri
      }
      env {
        name  = "NEO4J_USER"
        value = var.neo4j_user
      }
      env {
        name  = "NEO4J_PASSWORD"
        value = var.neo4j_password
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.run]
}

# Payment Service
resource "google_cloud_run_v2_service" "payment_service" {
  name     = "payment-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/payment-service:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "MONGODB_URI"
        value = var.mongodb_uri
      }
      env {
        name  = "TOUR_SERVICE_URL"
        value = google_cloud_run_v2_service.tour_service.uri
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.run]
}

# Blog Service
resource "google_cloud_run_v2_service" "blog_service" {
  name     = "blog-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/blog-service:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "SPRING_DATA_MONGODB_URI"
        value = "${var.mongodb_uri_blog}"
      }
      env {
        name  = "SERVER_PORT"
        value = "8080"
      }
      env {
        name  = "BASE_URL"
        value = ""  # Will be updated after deployment
      }

      resources {
        limits = {
          memory = "1Gi"
          cpu    = "1"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [google_project_service.run]
}

# Stakeholders Service
resource "google_cloud_run_v2_service" "stakeholders_service" {
  name     = "stakeholders-service"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/stakeholders-service:latest"

      ports {
        container_port = 8080
      }

      env {
        name  = "DB_HOST"
        value = "/cloudsql/${google_sql_database_instance.postgres.connection_name}"
      }
      env {
        name  = "DB_PORT"
        value = "5432"
      }
      env {
        name  = "DB_USER"
        value = "stakeholders_user"
      }
      env {
        name  = "DB_PASSWORD"
        value = var.postgres_password
      }
      env {
        name  = "DB_NAME"
        value = "stakeholders_db"
      }
      env {
        name  = "BASE_URL"
        value = ""  # Will be updated after deployment
      }

      resources {
        limits = {
          memory = "512Mi"
          cpu    = "1"
        }
      }
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.postgres.connection_name]
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  depends_on = [
    google_project_service.run,
    google_sql_database_instance.postgres,
    google_sql_database.stakeholders_db,
    google_sql_user.stakeholders_user
  ]
}

resource "google_cloud_run_v2_service_iam_member" "api_gateway_public" {
  name     = google_cloud_run_v2_service.api_gateway.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "tour_service_public" {
  name     = google_cloud_run_v2_service.tour_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "follower_service_public" {
  name     = google_cloud_run_v2_service.follower_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "payment_service_public" {
  name     = google_cloud_run_v2_service.payment_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "blog_service_public" {
  name     = google_cloud_run_v2_service.blog_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_cloud_run_v2_service_iam_member" "stakeholders_service_public" {
  name     = google_cloud_run_v2_service.stakeholders_service.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}


resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

data "google_project" "project" {
  project_id = var.project_id
}
