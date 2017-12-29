FROM docker.io/golang:1.9.2

ENV APP {{.APP}}
ENV PROJECT {{.PROJECT}}
ENV BASE_PATH {{.BASE_PATH}}
ENV LDFLAGS {{.LDFLAGS}}

copy src src

RUN cd $BASE_PATH; go install -ldflags "$LDFLAGS"
#RUN cd $BASE_PATH; go build -o $GOPATH/bin/$APP -ldflags "$LDFLAGS"
