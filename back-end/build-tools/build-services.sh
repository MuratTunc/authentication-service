#!/bin/bash

# Load .env variables if present
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

WAIT_TIME=${WAIT_TIME:-5}
ENVIRONMENT=${ENVIRONMENT:-production}
GOFULLPATH=${GOFULLPATH:-go}

function wait_for_services() {
  printf "⏳ Waiting for $WAIT_TIME seconds to allow services to initialize "
  for i in $(seq 1 $WAIT_TIME); do
    printf "."
    sleep 1
  done
  echo " ✅"
}

function build_service_binary() {
  local SERVICE_NAME=$1
  local SERVICE_PORT=$2
  local BINARY_NAME=$3

  echo "-->Cross-compiling $SERVICE_NAME binary for Linux..."
  echo "-->Stopping $SERVICE_NAME running Docker container on port $SERVICE_PORT..."
  container_id=$(docker ps --filter "publish=$SERVICE_PORT" -q)
  echo "-->Found container ID: $container_id"
  if [ -n "$container_id" ]; then
    docker stop $container_id && docker rm $container_id
    echo "✅ Stopped and removed container running on port $SERVICE_PORT."
  else
    echo "-->No running container found on port $SERVICE_PORT."
  fi

  echo "-->Cleaning up orphan containers (keeping volumes)..."
  docker-compose down --remove-orphans

  cd ../$SERVICE_NAME || exit 1
  mkdir -p bin
  GOOS=linux GOARCH=amd64 $GOFULLPATH build -o bin/$BINARY_NAME ./src/cmd
  echo "✅ Done! Linux binary created at ../$SERVICE_NAME/bin/$BINARY_NAME"
  cd - >/dev/null
}

function build_all() {
  build_service_binary "$AUTHENTICATION_SERVICE_NAME" "$AUTHENTICATION_SERVICE_PORT" "$AUTHENTICATION_SERVICE_BINARY"

  docker-compose up --build -d && \
  echo "✅ ✅ ✅ Docker images built and started!"

  wait_for_services

  echo "📜 Fetching logs for all services..."
  docker-compose logs --tail=20

  echo "✅ ✅ ✅ Running Containers:"
  docker ps

  run_integration_tests
}

function run_integration_tests() {
  echo "🔍 Running integration tests..."
  cd ../integration_tests || exit 1
  ./authentication-service.sh
  echo "✅ Integration tests completed successfully!"
}

function stop_down_containers() {
  if [ -z "$1" ]; then
    echo "🔍 Checking for running containers..."
    running_containers=$(docker ps -q)
    if [ -n "$running_containers" ]; then
      echo "🛑 Stopping all running Docker containers..."
      docker stop $running_containers
      echo "🗑️ Removing all stopped containers..."
      docker rm $(docker ps -aq)
      echo "🧹 Pruning unused Docker resources..."
      docker system prune -f
      echo "🛑 Stopping docker-compose..."
      docker-compose down
      echo "✅ All containers stopped, removed, and docker-compose down completed."
    else
      echo "⚡ No running containers found. Skipping stop and remove."
    fi
  else
    CONTAINER="$1"
    echo "🛑 Attempting to stop container: $CONTAINER"
    container_id=$(docker ps --filter "name=$CONTAINER" --format "{{.ID}}")
    if [ -n "$container_id" ]; then
      docker stop $container_id && docker rm $container_id
      echo "✅ Container $CONTAINER stopped and removed."
    else
      echo "⚠️ Container $CONTAINER not found or already stopped."
    fi
  fi
}

function print_vars() {
  echo "SERVER_IP = $SERVER_IP"
}

function show_logs() {
  echo "📜 Fetching last 20 logs for all services..."
  docker-compose logs --tail=20 -f
}

function usage() {
  echo "Usage: $0 [command]"
  echo "Commands:"
  echo "  build-all               Build and start all services"
  echo "  stop                    Stop and remove all containers"
  echo "  stop <container_name>   Stop and remove specific container"
  echo "  logs                    Show docker-compose logs"
  echo "  print-vars              Print environment variables"
  echo "  test                    Run integration tests"
}

# Entry point
case "$1" in
  build-all) build_all ;;
  stop) stop_down_containers "$2" ;;
  logs) show_logs ;;
  print-vars) print_vars ;;
  test) run_integration_tests ;;
  *) usage ;;
esac
