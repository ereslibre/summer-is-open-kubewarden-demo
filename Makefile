.PHONY: help
help:
	@echo "Two demos can be found here. You can run them with:"
	@echo "  - make kwctl-demo"
	@echo "  - make kubernetes-demo"

.PHONY: kwctl-demo
kwctl-demo: clear
	@go run . -0

.PHONY: kubernetes-demo
kubernetes-demo: clear
	@go run . -1

.PHONY: clear
clear:
	clear
