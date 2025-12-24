.PHONY: autofill resume cover-letter cold-message help

autofill:
	go run main.go

resume:
	go run ./cmd/resume/main.go

cover-letter:
	go run ./cmd/coverletter/main.go

cold-message:
	go run ./cmd/coldmessage/main.go

help:
	@echo "Available commands:"
	@echo "  make autofill      - Run the autofill application"
	@echo "  make resume        - Run the resume generator"
	@echo "  make cover-letter  - Generate cover letter"
	@echo "  make cold-message  - Generate cold outreach message"
