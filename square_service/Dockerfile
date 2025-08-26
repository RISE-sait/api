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

# Copy app code 
COPY --chown=appuser:appuser . /app

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Expose port (Cloud Run uses PORT env var)
EXPOSE 8080

# Run as non-root user
USER appuser

# Start the FastAPI app using our custom startup code
CMD ["python", "app.py"]
