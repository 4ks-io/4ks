.PHONY: plan apply validate-env

# make plan ENV=dev

# function to validate sbx,dev,prd argument value
validate-env:
	@if [ "$(ENV)" != "sbx" ] && [ "$(ENV)" != "dev" ] && [ "$(ENV)" != "prd" ]; then \
		echo "Error: ENV must be set to sbx, dev, or prd"; \
		exit 1; \
	fi

plan: validate-env
	terraform workspace select core-$(ENV)-us-east
	terraform plan -var-file=terraform.$(ENV).tfvars

apply: validate-env
	terraform workspace select core-$(ENV)-us-east
	terraform apply -var-file=terraform.$(ENV).tfvars

sbx:
	$(MAKE) apply ENV=sbx

dev:
	$(MAKE) apply ENV=dev

prd:
	$(MAKE) apply ENV=prd
