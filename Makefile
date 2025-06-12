# Nombre del binario base
BINARY_NAME=athena_server

# Versión del proyecto
VERSION=1.0.0

# Directorio de salida para los builds
BUILD_DIR=build

# Comando Go
GO=go

# ==============================================================================
# Comandos de Desarrollo
# ==============================================================================

.PHONY: start
start:
	@echo ">> Iniciando servidor de desarrollo..."
	$(GO) run ./cmd/server/main.go

.PHONY: test-unit
test-unit:
	@echo ">> Ejecutando pruebas unitarias..."
	$(GO) test -v -short ./...

.PHONY: test-integration
test-integration:
	@echo ">> Ejecutando pruebas de integración..."
	$(GO) test -v ./...

# ==============================================================================
# Comandos de Build
# ==============================================================================

.PHONY: build
build: clean-build
	@echo ">> Compilando binario para linux/amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64_v$(VERSION) ./cmd/server/main.go
	@echo "==> Binario generado en $(BUILD_DIR)/"

.PHONY: build-x86
build-x86: clean-build
	@echo ">> Compilando binario para linux/386 (32-bit)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GO) build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_386_v$(VERSION) ./cmd/server/main.go
	@echo "==> Binario generado en $(BUILD_DIR)/"

# ==============================================================================
# Limpieza
# ==============================================================================

.PHONY: clean
clean: clean-build
	@echo ">> Limpiando caché de Go..."
	$(GO) clean

# Nueva regla para limpiar SOLO el directorio de build
.PHONY: clean-build
clean-build:
	@echo ">> Limpiando directorio de build..."
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)