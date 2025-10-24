// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package runtime

/*
#cgo LDFLAGS: -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++ -fopenmp
#include <whisper.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <ggml-backend.h>

typedef int (*ggml_vk_get_device_count_t)(void);

static int whisper_vulkan_device_count() {
	void *handle = dlopen("libggml-vulkan.so", RTLD_LAZY | RTLD_LOCAL);
	if (handle == NULL) {
		return -1;
	}
	ggml_vk_get_device_count_t fn = (ggml_vk_get_device_count_t)dlsym(handle, "ggml_backend_vk_get_device_count");
	if (fn == NULL) {
		dlclose(handle);
		return -1;
	}
	int count = fn();
	dlclose(handle);
	return count;
}
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unsafe"

	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go"
)

func loadContext(path string, backend BackendType, gpuDevice int) (*whispercpp.Context, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ensureBackendsLoaded()

	params := C.whisper_context_default_params()

	switch backend {
	case BackendCPU:
		params.use_gpu = C.bool(false)
	case BackendVulkan:
		if C.whisper_vulkan_device_count() <= 0 {
			return nil, fmt.Errorf("no Vulkan devices detected")
		}
		params.use_gpu = C.bool(true)
		params.flash_attn = C.bool(false)
		if gpuDevice >= 0 {
			params.gpu_device = C.int(gpuDevice)
		} else {
			params.gpu_device = C.int(-1)
		}
	default:
		params.use_gpu = C.bool(false)
	}

	ctx := C.whisper_init_from_file_with_params(cPath, params)
	if ctx == nil {
		return nil, fmt.Errorf("%w: %s backend", ErrUnableToLoadModel, backend)
	}

	return (*whispercpp.Context)(ctx), nil
}

var backendRegistryOnce sync.Once

func ensureBackendsLoaded() {
	backendRegistryOnce.Do(func() {
		if dir := os.Getenv("GGML_BACKEND_PATH"); dir != "" {
			for _, entry := range strings.Split(dir, ":") {
				entry = strings.TrimSpace(entry)
				if entry == "" {
					continue
				}
				cDir := C.CString(entry)
				C.ggml_backend_load_all_from_path(cDir)
				C.free(unsafe.Pointer(cDir))
			}
		}
		C.ggml_backend_load_all()
	})
}
