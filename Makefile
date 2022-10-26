dsm-perf-tool: *.go cmd/*.go go.*
	go build

install: dsm-perf-tool
	sudo cp dsm-perf-tool /usr/local/bin
	@echo "Add the following line to your .bashrc file to enable bash completion:"
	@echo
	@echo "source <(dsm-perf-tool completion)"

clean:
	rm -f dsm-perf-tool
