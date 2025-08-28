.PHONY: test test-unit test-integration test-load coverage install-dev lint format

# Install development dependencies
install-dev:
	pip install -r requirements.dev.txt

# Run all tests except load tests
test:
	python -m pytest -m "not load" --cov=. --cov-report=html --cov-report=term-missing

# Run only unit tests
test-unit:
	python -m pytest tests/test_subscription_service.py tests/test_webhook_handlers.py -v

# Run only integration tests
test-integration:
	python -m pytest tests/test_integration.py -v

# Run load tests (requires running service)
test-load:
	python -m pytest tests/test_load.py -m load -v

# Generate coverage report
coverage:
	python -m pytest --cov=. --cov-report=html --cov-report=term-missing --cov-report=xml

# Run tests with coverage and open HTML report
coverage-html:
	python -m pytest --cov=. --cov-report=html
	@echo "Coverage report generated at htmlcov/index.html"

# Lint code
lint:
	flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics
	flake8 . --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics

# Format code
format:
	black . --line-length=127
	isort . --profile black

# Clean test artifacts
clean:
	rm -rf htmlcov/
	rm -rf .coverage
	rm -rf coverage.xml
	rm -rf .pytest_cache/
	rm -rf __pycache__/
	find . -name "*.pyc" -delete