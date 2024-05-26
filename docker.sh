docker build -f Dockerfile.dev -t docknat .
docker run -it --rm --cap-add=NET_ADMIN docknat sh
