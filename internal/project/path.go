package project

import (
	"path/filepath"
	"runtime"
)

var (
	// _, b, _, _ = runtime.Caller(0)
	// ProjectRoot es la ruta raíz del proyecto.
	// Se calcula subiendo dos niveles desde la ubicación de este archivo.
	// (desde /internal/project a /)
	// ProjectRoot = filepath.Join(filepath.Dir(b), "../..")

	_, b, _, _  = runtime.Caller(0)
	basepath    = filepath.Dir(b)
	ProjectRoot = filepath.Join(basepath, "..", "..")
)
