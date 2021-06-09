all: build

build:
	[ -f output ] || mkdir -p output
	go build -o output/c

clean:
	[ -f output ] && rm -r output