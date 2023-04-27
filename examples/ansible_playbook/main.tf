terraform {
  required_providers {
    ansible = {
      source  = "ansible/ansible"
      version = "~> 1.0.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }
}

# ===============================================
# Create a docker image using a Dockerfile
# ===============================================
resource "docker_image" "julia" {
  name = "julian-alps:latest"
  build {
    context    = "."
    dockerfile = "Dockerfile"
  }
}

