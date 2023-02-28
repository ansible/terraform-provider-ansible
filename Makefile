os = $(shell go env GOOS)
arch = $(shell go env GOARCH)

build-dev:
	go build -o ~/.terraform.d/plugins/terraform-ansible.com/ansibleprovider/ansible/0.0.2/$(os)_$(arch)/terraform-provider-ansible .
