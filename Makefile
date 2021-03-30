.PHONY: setup
setup:
	flutter channel master
	flutter upgrade
	flutter config --enable-windows-desktop
	flutter config --enable-macos-desktop
	flutter config --enable-linux-desktop
	sudo gem install cocoapods

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

BINDINGS_RELEASES         ?= https://github.com/nimona/go-nimona/releases/download
BINDINGS_VERSION          ?= v0.18.1
BINDINGS_ARTIFACT_ANDROID ?= /${BINDINGS_VERSION}/libnimona-${BINDINGS_VERSION}-android.tar.gz
BINDINGS_ARTIFACT_DARWIN  ?= /${BINDINGS_VERSION}/libnimona-${BINDINGS_VERSION}-darwin.tar.gz
BINDINGS_ARTIFACT_IOS     ?= /${BINDINGS_VERSION}/libnimona-${BINDINGS_VERSION}-ios.tar.gz
BINDINGS_ARTIFACT_LINUX   ?= /${BINDINGS_VERSION}/libnimona-${BINDINGS_VERSION}-linux.tar.gz
BINDINGS_ARTIFACT_WINDOWS ?= /${BINDINGS_VERSION}/libnimona-${BINDINGS_VERSION}-windows.tar.gz

.PHONY: upgrade-bindings
upgrade-bindings:
	mkdir -p android/src/main/jniLibs
	wget -c ${BINDINGS_RELEASES}${BINDINGS_ARTIFACT_ANDROID} -O - | tar -xz -C android/src/main/jniLibs
	mkdir -p macos
	wget -c ${BINDINGS_RELEASES}${BINDINGS_ARTIFACT_DARWIN} -O - | tar -xz -C macos
	mkdir -p ios
	wget -c ${BINDINGS_RELEASES}${BINDINGS_ARTIFACT_IOS} -O - | tar -xz -C ios
	mkdir -p linux/shared
	wget -c ${BINDINGS_RELEASES}${BINDINGS_ARTIFACT_LINUX} -O - | tar --strip-components=2 -xz -C linux/shared ./amd64
	mkdir -p windows/shared
	wget -c ${BINDINGS_RELEASES}${BINDINGS_ARTIFACT_WINDOWS} -O - | tar --strip-components=2 -xz -C windows/shared ./amd64
