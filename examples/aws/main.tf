terraform {
  required_providers {
    ansible = {
      version = "~> 0.0.1"
      source  = "terraform-ansible.com/ansibleprovider/ansible"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region     = "eu-north-1"
  access_key = "my_acces_key"
  secret_key = "my_secret_key"
}

# Add key for ssh connection
resource "aws_key_pair" "my_key" {
  key_name   = "my_key"
  public_key = "my_public_key_value"
}

# Add security group for ssh
resource "aws_security_group" "ssh" {
  name = "ssh"
  ingress {
    description = "ssh"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Add security group for http
resource "aws_security_group" "http" {
  name = "http"
  ingress {
    description = "http"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Set ami for ec2 instance
data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"]
}

# Create ec2 instance
resource "aws_instance" "my_ec2" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
  tags = {
    Name = "Inventory_plugin"
  }
  key_name        = aws_key_pair.my_key.key_name
  security_groups = [aws_security_group.ssh.name, aws_security_group.http.name, "default"]
}

# Add created ec2 instance to ansible inventory
resource "ansible_host" "my_ec2" {
  name   = aws_instance.my_ec2.public_dns
  groups = ["nginx"]
  variables = {
    ansible_user                 = "ubuntu",
    ansible_ssh_private_key_file = "~/.ssh/id_rsa",
    ansible_python_interpreter   = "/usr/bin/python3",
  }
}
