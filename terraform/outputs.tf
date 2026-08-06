output "api_gateway_url" {
  description = "API Gateway URL"
  value       = google_cloud_run_v2_service.api_gateway.uri
}

output "tour_service_url" {
  description = "Tour Service URL"
  value       = google_cloud_run_v2_service.tour_service.uri
}

output "follower_service_url" {
  description = "Follower Service URL"
  value       = google_cloud_run_v2_service.follower_service.uri
}

output "payment_service_url" {
  description = "Payment Service URL"
  value       = google_cloud_run_v2_service.payment_service.uri
}

output "blog_service_url" {
  description = "Blog Service URL"
  value       = google_cloud_run_v2_service.blog_service.uri
}

output "stakeholders_service_url" {
  description = "Stakeholders Service URL"
  value       = google_cloud_run_v2_service.stakeholders_service.uri
}

output "postgres_ip" {
  description = "Cloud SQL PostgreSQL public IP"
  value       = google_sql_database_instance.postgres.public_ip_address
}

output "postgres_connection_name" {
  description = "Cloud SQL connection name"
  value       = google_sql_database_instance.postgres.connection_name
}
