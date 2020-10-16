GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=./bin/mdldec
    
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps:
	$(GOGET) golang.org/x/image

all: build
