GOMOBILE_PKG := nimona.io/plugins/flutter
APP_PATH := $(CURDIR)

.PHONY: bind-ios
bind-ios:
	cd go; gomobile bind -v -target ios \
		-o ${APP_PATH}/plugins/identity_mobile/ios/Frameworks/Mobileapi.framework \
		nimona.io/plugins/flutter

.PHONY: bind-macos
bind-macos:
	cd go/binding; \
		go build \
		--buildmode=c-archive \
		-o ${APP_PATH}/macos/libnimona.a
