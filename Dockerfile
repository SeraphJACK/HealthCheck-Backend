FROM tetafro/golang-gcc:1.16-alpine AS build
COPY ./ /app/
WORKDIR /app/
ENV GOPROXY=https://goproxy.io,direct
RUN go build

FROM alpine:latest
WORKDIR /app
VOLUME /app/data
COPY --from=build /app/HealthCheck /app/
ENV GIN_MODE=release
ENTRYPOINT /app/HealthCheck --conf /app/data/config.yml --db /app/data/data.db
EXPOSE 8080/tcp
