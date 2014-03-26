.PHONY: apparatus-http-proxy
apparatus-http-proxy:
	go build

release: apparatus-http-proxy
	rm -rf apparatus-http-proxy.tgz
	tar zcvf apparatus-http-proxy.tgz apparatus-http-proxy start
