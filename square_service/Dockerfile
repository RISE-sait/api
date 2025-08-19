FROM python:3.11-slim

# Set working directory
WORKDIR /app



# Copy dependency list
COPY requirements.txt .


# Install build dependencies, install Python packages, then remove build deps
RUN apt-get update && apt-get install -y --no-install-recommends gcc libpq-dev \
    && pip install --no-cache-dir -r requirements.txt \
    && apt-get purge -y --auto-remove gcc libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -ms /bin/bash appuser

# Copy app code preserving package structure
COPY --chown=appuser:appuser . /app/square_service

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Expose port
EXPOSE 8000

# Run as non-root user
USER appuser

# Start the FastAPI app
CMD ["uvicorn", "square_service.app:app", "--host", "0.0.0.0", "--port", "8000", "--proxy-headers"]
