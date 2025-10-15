.PHONY: backend tidy

backend:
	cd backend && APP_AUTH_JWT_SECRET=dev-secret go run ./cmd/server

tidy:
	cd backend && go mod tidy
