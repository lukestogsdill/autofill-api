.PHONY: autofill resume help

autofill:
	go run main.go

resume:
	go run ./cmd/resume/main.go

help:
	@echo "Available commands:"
	@echo "  make autofill  - Run the autofill application"
	@echo "  make resume    - Run the resume generator"
