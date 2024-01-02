# Build docker image with dockerfile
```bash
$ docker build -t module3 .
```

# Run container with name module3
```bash
$ docker run -d --name module3 module3
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
