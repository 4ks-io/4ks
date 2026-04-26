.PHONY: plan apply validate-env work

plan:
	terraform workspace select media-sbx-us-east
	terraform plan

apply: validate-env
	terraform workspace select media-sbx-us-east
	terraform apply
