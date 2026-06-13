.PHONY: help backend-dev backend-build backend-test \
        backend-build-darwin-arm64 backend-build-linux-amd64 backend-build-windows-amd64 \
        frontend-install frontend-dev frontend-build \
        electron-install electron-dev \
        package-darwin package-linux package-win \
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
	@mkdir -p backend/bin/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -o backend/bin/darwin-arm64/gotutor-backend ./backend

backend-build-linux-amd64:
	@mkdir -p backend/bin/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o backend/bin/linux-amd64/gotutor-backend ./backend

backend-build-windows-amd64:
	@mkdir -p backend/bin/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o backend/bin/windows-amd64/gotutor-backend.exe ./backend

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

package-darwin: backend-build-darwin-arm64 frontend-build
	cd electron && pnpm exec electron-builder --mac --arm64

package-linux: backend-build-linux-amd64 frontend-build
	cd electron && pnpm exec electron-builder --linux --x64

package-win: backend-build-windows-amd64 frontend-build
	cd electron && pnpm exec electron-builder --win --x64

i18n-check:
	@echo "i18n key consistency check — implemented in Phase 12"

clean:
	rm -rf backend/bin frontend/dist electron/dist release
