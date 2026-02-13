APP_NAME   := Clicky
BUNDLE     := dist/$(APP_NAME).app
CONTENTS   := $(BUNDLE)/Contents
MACOS_DIR  := $(CONTENTS)/MacOS
RES_DIR    := $(CONTENTS)/Resources
ICON_SRC   := src/icon.png
ICONSET    := dist/icon.iconset

.PHONY: app windows clean

app: $(MACOS_DIR)/clicky $(CONTENTS)/Info.plist $(RES_DIR)/icon.icns

$(MACOS_DIR)/clicky: src/*.go src/objc_darwin.m src/objc_darwin.h
	@mkdir -p $(MACOS_DIR)
	cd src && CC=clang go build -o ../$(MACOS_DIR)/clicky

$(CONTENTS)/Info.plist: src/Info.plist
	@mkdir -p $(CONTENTS)
	cp src/Info.plist $(CONTENTS)/Info.plist

$(RES_DIR)/icon.icns: $(ICON_SRC)
	@mkdir -p $(RES_DIR) $(ICONSET)
	sips -z 16 16     $(ICON_SRC) --out $(ICONSET)/icon_16x16.png
	sips -z 32 32     $(ICON_SRC) --out $(ICONSET)/icon_16x16@2x.png
	sips -z 32 32     $(ICON_SRC) --out $(ICONSET)/icon_32x32.png
	sips -z 64 64     $(ICON_SRC) --out $(ICONSET)/icon_32x32@2x.png
	sips -z 128 128   $(ICON_SRC) --out $(ICONSET)/icon_128x128.png
	sips -z 256 256   $(ICON_SRC) --out $(ICONSET)/icon_128x128@2x.png
	sips -z 256 256   $(ICON_SRC) --out $(ICONSET)/icon_256x256.png
	sips -z 512 512   $(ICON_SRC) --out $(ICONSET)/icon_256x256@2x.png
	sips -z 512 512   $(ICON_SRC) --out $(ICONSET)/icon_512x512.png
	sips -z 1024 1024 $(ICON_SRC) --out $(ICONSET)/icon_512x512@2x.png
	iconutil -c icns $(ICONSET) -o $(RES_DIR)/icon.icns
	rm -rf $(ICONSET)

windows:
	cd src && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui" -o ../dist/clicky.exe

clean:
	rm -rf dist/
