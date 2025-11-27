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
	isSimulated := flag.Bool("s", false, "Simulated mode. No rename will take place.")
	flag.Parse()

	args := flag.Args()

	if len(args) < 3 {
		fmt.Println("Usage: rename [-s] <path> <in_regex_pattern> <out_template>")
		fmt.Println("Out Patterns:")
		fmt.Println("  :n:         Group n")
		fmt.Println("  :n,lower:   Group n lowercased")
		fmt.Println("  :n,upper:   Group n en upercased")
		fmt.Println("  :n,i:       Group n as integer number")
		fmt.Println("  :n,04i:     Group n zero padded with four zeros")
		fmt.Println("\nExample: rename -s ./pics \"img_(\\d+).jpg\" \"holidays_:1,04i:.jpg\"")
		os.Exit(1)
	}

	path := args[0]
	inRegexPattern := args[1]
	outTemplate := args[2]

	re, err := regexp.Compile(inRegexPattern)
	if err != nil {
		fmt.Printf("Error: Invalid regex: %v\n", err)
		os.Exit(1)
	}

	// 4. Leer el directorio
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("Error reading directory '%s': %v\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("Scanning directory: %s\n", path)
	if *isSimulated {
		fmt.Println("--- Simulated mode. No renames will take place ---")
	}

	count := 0
	errorCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		originalName := entry.Name()

		if re.MatchString(originalName) {
			// matches[0] is the whole filename, matches[1] is group 1, etc...
			matches := re.FindStringSubmatch(originalName)

			newName := buildNewName(outTemplate, matches)

			oldPath := filepath.Join(path, originalName)
			newPath := filepath.Join(path, newName)

			if originalName == newName {
				fmt.Printf("[=] %s\n", originalName)
			}

			if *isSimulated {
				fmt.Printf("[>] %s  ->  %s\n", originalName, newName)
			} else {
				err := os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Printf("[!] %s: %v\n", originalName, err)
					errorCount++
				} else {
					fmt.Printf("[>] %s  ->  %s\n", originalName, newName)
				}
			}
			count++
		}
	}

	if count == 0 {
		fmt.Printf("No files found with the pattern %s\n", inRegexPattern)
	} else if errorCount == 0 {
		fmt.Printf("%d files processed\n", count)
	} else {
		fmt.Printf("%d files processed and %d errors ocurred\n", count, errorCount)
	}
}

// buildNewName procesa el template buscando patrones :n: o :n,op1,op2:
func buildNewName(template string, matches []string) string {
	// Regex para encontrar :n: o :n,opciones:
	// Captura el número en grupo 1 y el resto (opciones) en el grupo 2
	placeholderRe := regexp.MustCompile(`:(\d+)((?:,[^:]+)*):`)

	result := placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		// match es algo como ":1:" o ":1,lower,03i:"

		// Quitamos los dos puntos de los extremos
		content := match[1 : len(match)-1]

		// Separamos por comas para obtener [indice, op1, op2...]
		parts := strings.Split(content, ",")

		// Parsear índice
		idxStr := parts[0]
		idx, err := strconv.Atoi(idxStr)

		// Si el índice no es válido, devolvemos el texto original sin tocar
		if err != nil || idx < 0 || idx >= len(matches) {
			return match
		}

		// Valor inicial capturado
		val := matches[idx]

		// Aplicar operaciones en orden
		for _, op := range parts[1:] {
			val = applyOperation(val, op)
		}

		return val
	})

	return result
}

// applyOperation aplica una transformación específica al valor
func applyOperation(val, op string) string {
	switch op {
	case "lower":
		return strings.ToLower(val)
	case "upper":
		return strings.ToUpper(val)
	case "i":
		// Convertir a entero simple (quita ceros a la izquierda)
		if num, err := strconv.Atoi(val); err == nil {
			return strconv.Itoa(num)
		}
	default:
		// Detectar patrón de padding: empieza con 0, termina con i (ej: 04i)
		if strings.HasPrefix(op, "0") && strings.HasSuffix(op, "i") {
			// Extraer el ancho (lo que hay entre 0 e i)
			widthStr := op[1 : len(op)-1]
			width, err := strconv.Atoi(widthStr)

			// Si el ancho es válido y el valor es un número
			if err == nil {
				if num, err := strconv.Atoi(val); err == nil {
					// Crear formato dinámico, ej: "%04d"
					format := fmt.Sprintf("%%0%dd", width)
					return fmt.Sprintf(format, num)
				}
			}
		}
	}
	// Si no se reconoce la operación o falla la conversión, devolver original
	return val
}
