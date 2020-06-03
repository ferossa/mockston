TARGETS = mockston

BUILD_LIST = $(foreach int, $(TARGETS), $(int)_build)

.PHONY: $(BUILD_LIST)

all: build

build: $(BUILD_LIST)

$(BUILD_LIST): %_build:
	go build -o bin/$* github.com/ferossa/mockston/cmd/$*