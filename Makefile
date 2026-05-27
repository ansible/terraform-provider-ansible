build:
	go build -o terraform-provider-ansible

legacy-tests: build
	cd tests/terraform_tests && ./run_tftest.sh

unit-tests:
	go test -v -count=1 ./framework/ -run "^Test[^A]"

acc-tests:
	TF_ACC=1 go test -v -count=1 -tags integration ./framework/ -run "^TestAcc"

tests: legacy-tests unit-tests acc-tests
