.PHONY: plan apply validate-env work


work:
	terraform workspace list


# function to validate sbx,dev,prd argument value
validate-env:
	@if [ "$(ENV)" != "sbx" ] && [ "$(ENV)" != "dev" ] && [ "$(ENV)" != "prd" ]; then \
		echo "Error: ENV must be set to sbx, dev, or prd"; \
		exit 1; \
	fi

plan: validate-env
	terraform workspace select base-$(ENV)-us-east
	terraform plan

apply: validate-env
	terraform workspace select base-$(ENV)-us-east
	terraform apply

sbx:
	$(MAKE) apply ENV=sbx

dev:
	$(MAKE) apply ENV=dev

prd:
	$(MAKE) apply ENV=prd
