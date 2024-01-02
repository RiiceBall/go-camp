# Clone project and build an executable http-server
FROM golang:1.16-alpine AS build
RUN apk add --no-cache git
RUN git clone https://github.com/RiiceBall/go-camp go-camp
RUN go build -o /bin/http-server ./go-camp/module2/http-server.go

# Copy executable file and run
FROM alpine
COPY --from=build /bin/http-server /bin/http-server
EXPOSE 8080
ENTRYPOINT ["/bin/http-server"]
