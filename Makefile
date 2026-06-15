.PHONY: help backend-dev backend-build backend-test \
        backend-build-darwin-arm64 backend-build-darwin-x64 \
        backend-build-linux-amd64 backend-build-windows-amd64 \
        frontend-install frontend-dev frontend-build \
        electron-install electron-dev electron-build \
        package-darwin package-darwin-x64 package-linux package-win \
        i18n-check clean

help:
	@echo "GoTutor — available targets:"
	@echo ""
	@echo "Development:"
	@echo "  backend-dev        Run Go backend with hot reload on :8081"
	@echo "  frontend-dev       Run Vite dev server on :5173"
	@echo "  electron-dev       Launch Electron (wraps frontend + spawns backend)"
	@echo ""
	@echo "Build:"
	@echo "  backend-build      Cross-compile Go backend to backend/bin/<os>-<arch>/"
	@echo "  backend-test       Run go test ./... in backend/"
	@echo "  frontend-build     Production Vite build → frontend/dist/"
	@echo ""
	@echo "Package (requires backend-build + frontend-build):"
	@echo "  package-darwin     macOS dmg (arm64)"
	@echo "  package-linux      Linux AppImage + deb (amd64)"
	@echo "  package-win        Windows nsis installer (amd64)"
	@echo ""
	@echo "Misc:"
	@echo "  i18n-check         Verify zh-CN and en locale keys match"
	@echo "  clean              Remove build artifacts"

backend-dev:
	cd backend && go run . -port 8081

backend-build: backend-build-darwin-arm64 backend-build-linux-amd64 backend-build-windows-amd64

backend-build-darwin-arm64:
	@mkdir -p backend/bin/mac-arm64
	cd backend && GOOS=darwin GOARCH=arm64 go build -o bin/mac-arm64/gotutor-backend .

backend-build-darwin-x64:
	@mkdir -p backend/bin/mac-x64
	cd backend && GOOS=darwin GOARCH=amd64 go build -o bin/mac-x64/gotutor-backend .

backend-build-linux-amd64:
	@mkdir -p backend/bin/linux-x64
	cd backend && GOOS=linux GOARCH=amd64 go build -o bin/linux-x64/gotutor-backend .

backend-build-windows-amd64:
	@mkdir -p backend/bin/win-x64
	cd backend && GOOS=windows GOARCH=amd64 go build -o bin/win-x64/gotutor-backend.exe .

backend-test:
	cd backend && go test ./...

frontend-install:
	cd frontend && pnpm install

frontend-dev:
	cd frontend && pnpm dev

frontend-build:
	cd frontend && pnpm build

electron-install:
	cd electron && pnpm install

electron-dev:
	cd electron && pnpm dev

electron-build:
	cd electron && pnpm build

package-darwin: backend-build-darwin-arm64 frontend-build electron-build
	cd electron && node scripts/prebuild.js --mac --arm64 && pnpm exec electron-builder --mac --arm64

package-darwin-x64: backend-build-darwin-x64 frontend-build electron-build
	cd electron && node scripts/prebuild.js --mac --x64 && pnpm exec electron-builder --mac --x64

package-linux: backend-build-linux-amd64 frontend-build electron-build
	cd electron && node scripts/prebuild.js --linux --x64 && pnpm exec electron-builder --linux --x64

package-win: backend-build-windows-amd64 frontend-build electron-build
	cd electron && node scripts/prebuild.js --win --x64 && pnpm exec electron-builder --win --x64

i18n-check:
	@echo "i18n key consistency check — implemented in Phase 12"

clean:
	rm -rf backend/bin frontend/dist electron/dist release
