terraform {
  required_providers {
    ansible = {
      version = "~> 1.0.0"
      source  = "ansible/ansible"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }
}

provider "docker" {
}

# ===============================================
# Create a docker container to use as host
# ===============================================
resource "docker_image" "alpine" {
  name = "alpine:latest"
  build {
    context    = "."
    dockerfile = "Dockerfile"
  }
}

resource "docker_container" "alpine" {
  image             = docker_image.alpine.image_id
  name              = "alpine-docker"
  must_run          = true
  publish_all_ports = true

  command = [
    "sleep",
    "infinity"
  ]
}


# ===============================================
# Run ansible playbook "example-play.yml" on a previously created host
# ===============================================
resource "ansible_provision" "provision" {
  playbook           = "example-play.yml"
  hostname           = docker_container.alpine.name
  hostgroup          = "provision"
  port               = 8080
  ansible_connection = "docker" # use "docker" if you're using a docker container as a host

  replayable = true
  verbosity  = 2
}