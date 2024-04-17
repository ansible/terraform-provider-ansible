build:
	go build -o terraform-provider-ansible

testacc:
	TF_ACC=1 go test -v ./...
