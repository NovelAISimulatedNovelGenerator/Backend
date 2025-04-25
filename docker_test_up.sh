docker-compose down --volumes --remove-orphans
docker system prune -af
docker volume prune -f
docker rmi $(docker images -q novelai_app) || true
docker-compose build --no-cache
docker-compose up