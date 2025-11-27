package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// 1. Definir el flag opcional -n para "dry-run" (simulación)
	dryRun := flag.Bool("n", false, "Modo simulación: muestra qué pasaría sin renombrar nada")
	flag.Parse()

	// 2. Obtener los argumentos posicionales (después de los flags)
	args := flag.Args()

	// Validar que tenemos los 3 argumentos requeridos
	if len(args) < 3 {
		fmt.Println("Uso: go run renamer.go [-n] <ruta> <regex> <nuevo_patron>")
		fmt.Println("Ejemplo: go run renamer.go -n ./mis_fotos \"img_(\\d+).jpg\" \"vacaciones_/1.jpg\"")
		os.Exit(1)
	}

	dirPath := args[0]
	regexPattern := args[1]
	replacementTemplate := args[2]

	// 3. Compilar la expresión regular
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Printf("Error: La expresión regular no es válida: %v\n", err)
		os.Exit(1)
	}

	// 4. Leer el directorio
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error al leer el directorio '%s': %v\n", dirPath, err)
		os.Exit(1)
	}

	fmt.Printf("Escaneando directorio: %s\n", dirPath)
	if *dryRun {
		fmt.Println("--- MODO SIMULACIÓN (No se harán cambios) ---")
	}

	count := 0
	for _, entry := range entries {
		// Ignoramos directorios, solo renombramos archivos
		if entry.IsDir() {
			continue
		}

		originalName := entry.Name()

		// 5. Verificar si el archivo coincide con la regex
		if re.MatchString(originalName) {
			// Obtener los grupos capturados (matches)
			// matches[0] es todo el string, matches[1] es el grupo 1, etc.
			matches := re.FindStringSubmatch(originalName)

			// 6. Generar el nuevo nombre reemplazando /1, /2, etc.
			newName := buildNewName(replacementTemplate, matches)

			// Construir rutas completas
			oldPath := filepath.Join(dirPath, originalName)
			newPath := filepath.Join(dirPath, newName)

			// Evitar renombrar si el nombre no cambia
			if originalName == newName {
				continue
			}

			// 7. Ejecutar renombrado o simulación
			if *dryRun {
				fmt.Printf("[Simul.] %s  ->  %s\n", originalName, newName)
			} else {
				err := os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Printf("[Error] No se pudo renombrar %s: %v\n", originalName, err)
				} else {
					fmt.Printf("[OK] %s  ->  %s\n", originalName, newName)
				}
			}
			count++
		}
	}

	if count == 0 {
		fmt.Println("No se encontraron archivos que coincidan con el patrón.")
	} else {
		fmt.Printf("Proceso finalizado. %d archivos procesados.\n", count)
	}
}

// buildNewName toma el template (ej: "foto_/1.jpg") y el slice de matches,
// y retorna el string final reemplazando /n por el contenido del grupo.
func buildNewName(template string, matches []string) string {
	// Esta regex busca patrones literales "/numero" en el template de reemplazo
	// Ejemplo: busca "/1", "/20", etc.
	placeholderRe := regexp.MustCompile(`/(\d+)`)

	result := placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		// match es algo como "/1"
		// Quitamos la barra para obtener el número
		idxStr := strings.TrimPrefix(match, "/")
		idx, err := strconv.Atoi(idxStr)

		// Si no es un número válido o el índice está fuera de rango,
		// devolvemos el match original tal cual (para no romper nada).
		if err != nil || idx < 0 || idx >= len(matches) {
			return match
		}

		// Retornamos el contenido capturado por la regex original
		return matches[idx]
	})

	return result
}
