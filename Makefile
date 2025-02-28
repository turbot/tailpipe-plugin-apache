TAILPIPE_INSTALL_DIR ?= ~/.tailpipe
BUILD_TAGS = netgo
install:
	go build -o $(TAILPIPE_INSTALL_DIR)/plugins/hub.tailpipe.io/plugins/turbot/apache@latest/tailpipe-plugin-apache.plugin -tags "${BUILD_TAGS}" *.go