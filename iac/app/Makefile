.PHONY: plan apply validate-env work


work:
	terraform workspace list


# function to validate dev or prd argument value
validate-env:
	@if [ "$(ENV)" != "dev" ] && [ "$(ENV)" != "prd" ]; then \
		echo "Error: ENV must be set to dev, or prd"; \
		exit 1; \
	fi

plan: validate-env
	terraform workspace select app-$(ENV)-us-east
	terraform plan -var-file=terraform.$(ENV).tfvars

apply: validate-env
	terraform workspace select app-$(ENV)-us-east
	terraform apply -var-file=terraform.$(ENV).tfvars

dev:
	$(MAKE) apply ENV=dev

prd:
	$(MAKE) apply ENV=prd
