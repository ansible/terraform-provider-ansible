build:
	go build -o terraform-provider-ansible

test: build
	cd tests/terraform_tests && ./run_tftest.sh
