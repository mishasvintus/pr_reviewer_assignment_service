build: 
	go build -o bin/api ./cmd/api

run: 
	go run ./cmd/api

test-unit: ## Запустить unit-тесты
	go test -v ./tests/unit_tests/...

test-integration: ## Запустить integration-тесты (требует запущенную БД)
	@echo "Убедитесь, что PostgreSQL запущен: docker-compose up -d postgres"
	go test -v ./tests/integration/...

test-all: ## Запустить все тесты (unit + integration)
	go test -v ./tests/unit_tests/... ./tests/integration/...

test-coverage: ## Показать покрытие тестами
	go test -coverprofile=coverage.out -coverpkg=./internal/... ./tests/unit_tests/... ./tests/integration/...
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет сохранен в coverage.html"

loadtest-burst: ## Запустить burst нагрузочное тестирование
	@echo "Убедитесь, что сервис запущен на http://localhost:8080"
	go test -v ./tests/stress_tests/... -run TestBurstLoadTest

loadtest-rampup: ## Запустить ramp-up нагрузочное тестирование
	@echo "Убедитесь, что сервис запущен на http://localhost:8080"
	go test -v ./tests/stress_tests/... -run TestRampUpReassignPR

loadtest-all: ## Запустить все нагрузочные тесты
	@echo "Убедитесь, что сервис запущен на http://localhost:8080"
	go test -v ./tests/stress_tests/...

fmt: ## Отформатировать код с помощью gofmt
	go fmt ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-up: 
	docker-compose up -d

docker-down:
	docker-compose down

