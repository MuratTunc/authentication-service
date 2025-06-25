#!/bin/bash

echo "Stopping all running Docker containers..."
docker ps -q | xargs -r docker stop

echo "Removing all containers..."
docker ps -aq | xargs -r docker rm

echo "Removing all Docker images..."
docker images -q | xargs -r docker rmi -f

echo "All containers and images have been removed."
