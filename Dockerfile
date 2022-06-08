FROM alpine:3.16.0
ARG GOLANG_VERSION=1.18.2

ENV APPMAN_LOGIN_HOST=localhost 
ENV APPMAN_LOGIN_PORT=7043
ENV APPMAN_DATABASE_HOST=localhost
ENV APPMAN_DATABASE_PORT=5432
ENV APPMAN_DATABASE_USERNAME=postgres
ENV APPMAN_DATABASE_PASSWORD=password
ENV APPMAN_DATABASE_NAME=application_service

# Install required packages
RUN apk update
RUN apk add go gcc bash musl-dev openssl-dev ca-certificates
RUN update-ca-certificates

# Install go
RUN wget https://dl.google.com/go/go$GOLANG_VERSION.src.tar.gz
RUN tar -C /usr/local -xzf go$GOLANG_VERSION.src.tar.gz
RUN cd /usr/local/go/src && ./make.bash
ENV PATH=$PATH:/usr/local/go/bin
RUN rm -f go$GOLANG_VERSION.src.tar.gz
RUN apk del go

COPY . .
RUN go build -o build/server ./src/main.go

CMD [ "build/server" ]
EXPOSE 7043