default: testacc docs build

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: docs
docs:
	go generate

.PHONY: build
build:
	go build -o terraform-provider-ansible
