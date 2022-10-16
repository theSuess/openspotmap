FROM docker.io/golang:1.18 AS pkgbuild
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o openspotmap .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=pkgbuild /workspace/openspotmap .
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/openspotmap"]
