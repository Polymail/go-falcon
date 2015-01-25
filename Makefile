ifndef FALCONGOBIN
  FALCONGOBIN=go
endif

all: deps
		$(FALCONGOBIN) install

deps:
		$(FALCONGOBIN) get launchpad.net/goyaml
		$(FALCONGOBIN) get github.com/lib/pq
		$(FALCONGOBIN) get golang.org/x/text/encoding
		$(FALCONGOBIN) get golang.org/x/text/transform
		$(FALCONGOBIN) get github.com/garyburd/redigo/redis
		$(FALCONGOBIN) get github.com/sloonz/go-iconv
		$(FALCONGOBIN) get github.com/sloonz/go-qprintable
		$(FALCONGOBIN) get launchpad.net/gocheck
test:
		$(FALCONGOBIN) test -v ./...
