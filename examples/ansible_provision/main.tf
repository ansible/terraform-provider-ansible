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
#    aws = {
#      source = "hashicorp/aws"
#      version = "~> 4.0"
#    }
  }
}

provider "docker" {
}

#provider "aws" {
#  access_key = "ansible_key"
#  secret_key = "secret_key"
#  region = "region"
#}

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


# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
# PROVISIONING AWS UBUNTU VIRTUAL MACHINE
# ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
#
# ==========
# Variables
# ==========
#variable "network_interface_id" {
#  type = string
#  default = "eni-00cd1bffc63dc7384"
#}
#
#variable "aws_hostname" {
#  type = string
#  default = "aws-alpine-host"
#}
#
#resource "tls_private_key" "private_key" {
#  algorithm = "RSA"
#  rsa_bits = 4096
#}
#
#resource "aws_key_pair" "ssh_key" {
#  key_name = "ssh_key"
#  public_key = tls_private_key.private_key.public_key_openssh
#
#  provisioner "local-exec" {
#    command = "echo '${tls_private_key.private_key.private_key_pem}' > ./'${aws_key_pair.ssh_key.key_name}'.pem"
#  }
#}
#
#resource "aws_security_group" "security_group" {
#  name_prefix = "aws-"
#  ingress {
#    from_port = 0
#    protocol  = "tcp"
#    to_port   = 0
#    cidr_blocks = ["0.0.0.0/0"]
#  }
#
#  egress {
#    from_port = 0
#    protocol  = "tcp"
#    to_port   = 0
#    cidr_blocks = ["0.0.0.0/0"]
#  }
#}
#
#resource "aws_instance" "aws" {
#  ami = "ami-0ec7f9846da6b0f61"  # alpine linux 3.17.1 ami
#  instance_type = "t2.micro"
#  key_name = aws_key_pair.ssh_key.key_name
#  vpc_security_group_ids = [aws_security_group.security_group.id]
#
#  tags = {
#    Name = var.aws_hostname
#  }
#
##  network_interface {
##    device_index         = 0
##    network_interface_id = var.network_interface_id
##  }
#
#  # Set hosntame to aws-alpine-host then, install python on it
#  user_data = <<EOF
##!/bin/sh
## set hostname
#hostnamectl set-hostname ${var.aws_hostname}
#
## install python
##  apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
##  python3 -m ensurepip
##  pip3 install --no-cache --upgrade pip setuptools
##  sleep infinity
#EOF
#
#  credit_specification {
#    cpu_credits = "unlimited"
#  }
#}
#
#resource "ansible_provision" "provision" {
#  playbook           = "example-play.yml"
#  hostname           = aws_instance.aws.public_ip
#  hostgroup          = "provision"
#  port               = 8080
#  remote_user = "ubuntu"
#  ansible_connection = "aws" # use "aws" if you're using an aws instance as a host
#  extra_vars = {
#    ansible_ssh_private_key_file = "./ssh_key.pem"
#  }
#
#  replayable = true
#  verbosity  = 2
#
#}
