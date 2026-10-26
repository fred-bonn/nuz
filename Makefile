OUT_PATH=./bin/nuz
TEST_PKG ?= ./...
TEST_ARGS ?=
SERVE_ARGS ?=

all: run

test:
	go test $(TEST_PKG) $(TEST_ARGS)

build: test
	go build -o $(OUT_PATH) ./cmd/cli

run: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v 

brief: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt

rain: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v -w 1

sun: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v -w 2

sandstorm: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v -w 3

hail: build
	$(OUT_PATH) showdown_demo_files/player.txt showdown_demo_files/opponent.txt -v -w 4

serve: test
	go run ./cmd/server $(SERVE_ARGS)

ITERATION_GOALS := $(filter-out all test build run serve brief rain sun sandstorm hail,$(MAKECMDGOALS))

.PHONY: $(ITERATION_GOALS)
$(ITERATION_GOALS): build
	@case "$@" in *[!0-9]*|0) echo "usage: make <positive iteration count>"; exit 1;; esac
	$(OUT_PATH) data/player.txt data/opponent.txt --iterations $@
