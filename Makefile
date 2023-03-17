os = $(shell go env GOOS)
arch = $(shell go env GOARCH)

build-dev:
	go build -o ~/.terraform.d/plugins/registry.terraform.io/ansible/ansible/1.0.0/$(os)_$(arch)/terraform-provider-ansible .
