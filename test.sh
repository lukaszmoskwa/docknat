docker build -f Dockerfile.dev -t docknat .
docker run -it --rm --cap-add=NET_ADMIN docknat sh -c "cd .. && go test ./... -v -p 1 -count=1 -coverprofile=coverage.out"
