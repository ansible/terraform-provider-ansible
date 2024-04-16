build:
	go build -o terraform-provider-ansible

test: build
	cd tests/terraform_tests && ./run_tftest.sh

testacc:
	@echo "==> Running acceptance tests..."
	TF_ACC=1 go test -v ./...
