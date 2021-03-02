GOMOBILE_PKG := nimona.io/plugins/flutter
APP_PATH := $(CURDIR)

.PHONY: bind-ios
bind-ios:
	cd go; gomobile bind -v -target ios \
		-o ${APP_PATH}/plugins/identity_mobile/ios/Frameworks/Mobileapi.framework \
		nimona.io/plugins/flutter

.PHONY: bind
bind: bind-macos bind-ios

.PHONY: bind-ios
bind-ios:
	cd go/binding; \
		BINDING_ARGS="-tags ios" \
		CGO_CFLAGS="-fembed-bitcode" \
		CGO_ENABLED=1 \
		GOARCH=arm64 \
		GOOS=darwin \
		SDK=iphoneos CC=$(PWD)/clangwrap.sh \
		go build \
		--buildmode=c-archive \
		-o ${APP_PATH}/ios/libnimona.a

.PHONY: bind-macos
bind-macos:
	cd go/binding; \
		go build \
		--buildmode=c-archive \
		-o ${APP_PATH}/macos/libnimona.a
