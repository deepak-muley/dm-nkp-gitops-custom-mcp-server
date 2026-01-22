# Dockerfile for dm-nkp-gitops-kmcp-server
# K8s-native Python MCP server using FastMCP

FROM python:3.12-slim as builder

WORKDIR /app

# Install build dependencies
RUN pip install --no-cache-dir hatchling

# Copy project files
COPY pyproject.toml .
COPY src/ src/

# Build wheel
RUN pip wheel --no-cache-dir --wheel-dir /wheels .

# Production image
FROM python:3.12-slim

# Security: Run as non-root
RUN useradd --create-home --uid 1000 appuser

WORKDIR /app

# Install the wheel and dependencies
COPY --from=builder /wheels /wheels
RUN pip install --no-cache-dir /wheels/*.whl && rm -rf /wheels

# Switch to non-root user
USER appuser

# Expose port (for A2A HTTP mode)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD python -c "import socket; s=socket.socket(); s.connect(('localhost', 8080)); s.close()" || exit 1

# Run the MCP server
ENTRYPOINT ["python", "-m", "gitops_mcp.server"]
