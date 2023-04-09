FROM golang as builder


RUN apt update && \
    apt install -y bash git make

WORKDIR /app
COPY . .

ENV CGO_ENABLED=1

RUN go mod tidy
RUN go build -tags netgo -a -v -installsuffix cgo -o bin/minion main.go 


FROM ubuntu
RUN apt update \
    && apt install -y curl wget \
    && apt install -y ca-certificates \
    && update-ca-certificates 2>/dev/null || true

COPY --from=builder /app/bin/minion /minion

CMD ["/minion"]