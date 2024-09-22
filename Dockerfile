FROM public.ecr.aws/docker/library/golang:1.19-buster as build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o golang-sample

FROM public.ecr.aws/docker/library/debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/golang-sample /app/golang-sample

CMD ["/app/golang-sample"]
