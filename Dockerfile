FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /service
RUN go build -buildmode=plugin -o /module.so ./plugin/*.go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /service /service
COPY --from=build /module.so /module.so

EXPOSE 8880

USER nonroot:nonroot

ENTRYPOINT ["/service"]