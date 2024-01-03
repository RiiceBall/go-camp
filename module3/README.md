# Build docker image with dockerfile
```bash
$ docker build -t module3 .
```

# Run container with name module3
# Publish container 8080 port to host 8080 port
```bash
$ docker run -d -p 8080:8080 --name module3 module3
```

# Get container PID
```bash
$ docker inspect --format '{{ .State.Pid }}' module3
```

# Access to container net with nsenter (must have root permission)
```bash
$ sudo nsenter -t [CONTAINER_PID] --net
```

# Get container ip config
```bash
$ ip addr
```
