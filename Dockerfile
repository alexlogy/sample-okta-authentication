FROM golang:1.26.1-alpine as BUILD

LABEL AUTHOR=Alex \
      WEBSITE="https://alexlogy.io"

WORKDIR /app
COPY . .

RUN go mod download \
    && go mod verify \
    && CGO_ENABLED=0 GOOS=linux go build -o sample-okta-app

FROM alpine:3.23.3 as APP

LABEL AUTHOR=Alex \
      WEBSITE="https://alexlogy.io"

WORKDIR /app

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=BUILD /app/sample-okta-app /app/sample-okta-app

# Copy the Self-Signed Certificate and Key into the Docker Image
COPY *.crt /app
COPY *.key /app

# Copy Configuration into the Docker Image
COPY config.json /app

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Environment Variables
ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

USER appuser

ENTRYPOINT ["/app/sample-okta-app"]
