FROM docker.io/golang:1.9.2

copy src src

RUN cd {{.BASE_PATH}}; go install -ldflags {{.LDFLAGS}}
