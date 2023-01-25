# These are overridden by cloudbuild.yaml when run by Prow.

export REGISTRY ?= youngnick/delegated-client


# Prow gives this a value of the form vYYYYMMDD-hash.
# (It's similar to `git describe` output, and for non-tag
# builds will give vYYYYMMDD-COMMITS-HASH where COMMITS is the
# number of commits since the last tag.)
export GIT_TAG ?= dev

# Prow gives this the reference it's called on.
# The test-infra config job only allows our cloudbuild to
# be called on `main` and semver tags, so this will be
# set to one of those things.
export BASE_REF ?= main

# The commit hash of the current checkout
# Used to pass a binary version for main,
# overridden to semver for tagged versions.
# Cloudbuild will set this in the environment to the
# commit SHA, since the Prow does not seem to check out
# a git repo.
export COMMIT ?= $(shell git rev-parse --short HEAD)


all: bin/dclient bin/sclient

bin/dclient: cmd/dclient/main.go
	go build -o bin/dclient ./cmd/dclient

bin/sclient: cmd/sclient/main.go
	go build -o bin/sclient ./cmd/sclient

.PHONY: push

push:
	./hack/build-and-push.sh