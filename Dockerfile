FROM golang:1 as builder
COPY . /src
WORKDIR /src
ENV CGO_ENABLED 0
RUN go get -d ./...
RUN go build -o /assets/in ./cmd/in
RUN go build -o /assets/out ./cmd/out
RUN go build -o /assets/check ./cmd/check
RUN set -e; for pkg in $(go list ./...); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM alpine:edge AS resource
RUN apk add --no-cache bash tzdata ca-certificates git jq openssh
RUN git config --global user.email "git@localhost"
RUN git config --global user.name "git"
COPY --from=builder assets/ /opt/resource/
RUN chmod +x /opt/resource/*

FROM resource AS tests
ADD . /romver-resource
COPY --from=builder /tests /romver-resource/tests
ARG ROMVER_TESTING_GITHUB_URI
ARG ROMVER_TESTING_GITHUB_BRANCH
ARG ROMVER_TESTING_GITHUB_USERNAME
ARG ROMVER_TESTING_GITHUB_PASSWORD
WORKDIR /romver-resource/tests
RUN set -e; for test in ./*.test; do \
		$test -ginkgo.v; \
	done

FROM resource
