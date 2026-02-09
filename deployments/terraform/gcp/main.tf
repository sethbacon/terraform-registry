terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  # Uncomment for remote state:
  # backend "gcs" {
  #   bucket = "my-terraform-state"
  #   prefix = "terraform-registry"
  # }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# ---------------------------------------------------------------------------
# Enable APIs
# ---------------------------------------------------------------------------
resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "secretmanager.googleapis.com",
    "artifactregistry.googleapis.com",
    "vpcaccess.googleapis.com",
    "compute.googleapis.com",
  ])

  project = var.project_id
  service = each.value

  disable_on_destroy = false
}

# ---------------------------------------------------------------------------
# VPC & Connector
# ---------------------------------------------------------------------------
resource "google_compute_network" "main" {
  name                    = "${var.name}-vpc"
  auto_create_subnetworks = false

  depends_on = [google_project_service.apis]
}

resource "google_compute_subnetwork" "main" {
  name          = "${var.name}-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.main.id
}

resource "google_compute_global_address" "private_ip" {
  name          = "${var.name}-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.main.id
}

resource "google_service_networking_connection" "private" {
  network                 = google_compute_network.main.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip.name]
}

resource "google_vpc_access_connector" "main" {
  name          = "${var.name}-conn"
  region        = var.region
  ip_cidr_range = "10.8.0.0/28"
  network       = google_compute_network.main.name
  machine_type  = "e2-micro"
  min_instances = 2
  max_instances = 3

  depends_on = [google_project_service.apis]
}

# ---------------------------------------------------------------------------
# Artifact Registry
# ---------------------------------------------------------------------------
resource "google_artifact_registry_repository" "main" {
  location      = var.region
  repository_id = var.name
  format        = "DOCKER"
  description   = "Terraform Registry container images"

  depends_on = [google_project_service.apis]
}

# ---------------------------------------------------------------------------
# Cloud SQL (PostgreSQL)
# ---------------------------------------------------------------------------
resource "google_sql_database_instance" "main" {
  name             = "${var.name}-db"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier              = var.db_tier
    availability_type = var.environment == "production" ? "REGIONAL" : "ZONAL"
    disk_size         = 20
    disk_type         = "PD_SSD"

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = var.environment == "production"
    }

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                               = google_compute_network.main.id
      enable_private_path_for_google_cloud_services = true
    }

    database_flags {
      name  = "max_connections"
      value = "100"
    }
  }

  deletion_protection = var.environment == "production"

  depends_on = [google_service_networking_connection.private]
}

resource "google_sql_database" "main" {
  name     = "terraform_registry"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "main" {
  name     = "registry"
  instance = google_sql_database_instance.main.name
  password = var.database_password
}

# ---------------------------------------------------------------------------
# GCS Storage Bucket
# ---------------------------------------------------------------------------
resource "google_storage_bucket" "storage" {
  name     = "${var.name}-storage-${var.project_id}"
  location = var.region

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 5
    }
    action {
      type = "Delete"
    }
  }
}

# ---------------------------------------------------------------------------
# Secret Manager
# ---------------------------------------------------------------------------
resource "google_secret_manager_secret" "db_password" {
  secret_id = "${var.name}-db-password"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.database_password
}

resource "google_secret_manager_secret" "jwt_secret" {
  secret_id = "${var.name}-jwt-secret"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "jwt_secret" {
  secret      = google_secret_manager_secret.jwt_secret.id
  secret_data = var.jwt_secret
}

resource "google_secret_manager_secret" "encryption_key" {
  secret_id = "${var.name}-encryption-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "encryption_key" {
  secret      = google_secret_manager_secret.encryption_key.id
  secret_data = var.encryption_key
}

# ---------------------------------------------------------------------------
# Service Account
# ---------------------------------------------------------------------------
resource "google_service_account" "backend" {
  account_id   = "${var.name}-backend"
  display_name = "Terraform Registry Backend"
}

resource "google_project_iam_member" "backend_sql" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "backend_secrets" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_storage_bucket_iam_member" "backend_storage" {
  bucket = google_storage_bucket.storage.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.backend.email}"
}

# ---------------------------------------------------------------------------
# Cloud Run - Backend
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "backend" {
  name     = "${var.name}-backend"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"

  template {
    service_account = google_service_account.backend.email

    scaling {
      min_instance_count = var.backend_min_instances
      max_instance_count = var.backend_max_instances
    }

    vpc_access {
      connector = google_vpc_access_connector.main.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.main.connection_name]
      }
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.name}/backend:${var.image_tag}"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "1Gi"
        }
      }

      env {
        name  = "TFR_SERVER_HOST"
        value = "0.0.0.0"
      }
      env {
        name  = "TFR_SERVER_PORT"
        value = "8080"
      }
      env {
        name  = "TFR_SERVER_BASE_URL"
        value = var.domain != "" ? "https://${var.domain}" : ""
      }
      env {
        name  = "TFR_DATABASE_HOST"
        value = "/cloudsql/${google_sql_database_instance.main.connection_name}"
      }
      env {
        name  = "TFR_DATABASE_PORT"
        value = "5432"
      }
      env {
        name  = "TFR_DATABASE_NAME"
        value = "terraform_registry"
      }
      env {
        name  = "TFR_DATABASE_USER"
        value = "registry"
      }
      env {
        name  = "TFR_DATABASE_SSL_MODE"
        value = "disable"
      }
      env {
        name  = "TFR_SECURITY_TLS_ENABLED"
        value = "false"
      }
      env {
        name  = "TFR_STORAGE_DEFAULT_BACKEND"
        value = "gcs"
      }
      env {
        name  = "TFR_STORAGE_GCS_BUCKET"
        value = google_storage_bucket.storage.name
      }
      env {
        name  = "TFR_STORAGE_GCS_PROJECT_ID"
        value = var.project_id
      }
      env {
        name  = "TFR_AUTH_API_KEYS_ENABLED"
        value = "true"
      }
      env {
        name  = "TFR_LOGGING_LEVEL"
        value = "info"
      }
      env {
        name  = "TFR_LOGGING_FORMAT"
        value = "json"
      }
      env {
        name  = "TFR_TELEMETRY_ENABLED"
        value = "true"
      }
      env {
        name  = "TFR_TELEMETRY_METRICS_ENABLED"
        value = "true"
      }
      env {
        name  = "TFR_TELEMETRY_METRICS_PROMETHEUS_PORT"
        value = "9090"
      }
      env {
        name  = "DEV_MODE"
        value = "false"
      }

      env {
        name = "TFR_DATABASE_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.db_password.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "TFR_JWT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.jwt_secret.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "ENCRYPTION_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.encryption_key.secret_id
            version = "latest"
          }
        }
      }

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      startup_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 5
        period_seconds        = 5
        failure_threshold     = 10
      }

      liveness_probe {
        http_get {
          path = "/health"
          port = 8080
        }
        initial_delay_seconds = 10
        period_seconds        = 30
      }
    }
  }

  depends_on = [google_project_service.apis]
}

# ---------------------------------------------------------------------------
# Cloud Run - Frontend
# ---------------------------------------------------------------------------
resource "google_cloud_run_v2_service" "frontend" {
  name     = "${var.name}-frontend"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = var.frontend_min_instances
      max_instance_count = var.frontend_max_instances
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${var.name}/frontend:${var.image_tag}"

      ports {
        container_port = 80
      }

      resources {
        limits = {
          cpu    = "0.5"
          memory = "256Mi"
        }
      }

      startup_probe {
        http_get {
          path = "/"
          port = 80
        }
        initial_delay_seconds = 2
        period_seconds        = 3
        failure_threshold     = 5
      }

      liveness_probe {
        http_get {
          path = "/"
          port = 80
        }
        period_seconds = 30
      }
    }
  }

  depends_on = [google_project_service.apis]
}

# Allow public access to frontend
resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  project  = google_cloud_run_v2_service.frontend.project
  location = google_cloud_run_v2_service.frontend.location
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
