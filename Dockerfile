FROM public.ecr.aws/docker/library/golang:1.19-buster as build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o futures-core

FROM public.ecr.aws/docker/library/debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/futures-core /app/futures-core

CMD ["/app/futures-core"]
