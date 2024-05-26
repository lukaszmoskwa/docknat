docker build -f Dockerfile.dev -t docknat .
docker run -it --rm --cap-add=NET_ADMIN -v "/var/run/docker.sock:/var/run/docker.sock:rw" docknat sh
