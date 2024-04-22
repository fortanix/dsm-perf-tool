sdkms-perf-tool: *.go cmd/*.go go.*
	go build

install: sdkms-perf-tool
	sudo cp sdkms-perf-tool /usr/local/bin
	@echo "Add the following line to your .bashrc file to enable bash completion:"
	@echo
	@echo "source <(sdkms-perf-tool completion)"

clean:
	rm -f sdkms-perf-tool
