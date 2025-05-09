package Analyzer

import (
	"MIA_Proyecto1/backend/FileSystem"
	"MIA_Proyecto1/backend/Management"
	"MIA_Proyecto1/backend/User"
	"bufio" //Para leer la entrada del usuario
	"bytes"
	"flag"    //Para manejar parametros y opciones de comandos
	"fmt"     //imprimir
	"os"      // para ingre mediante consola
	"regexp"  //buscar y extraer parametros en la entrada
	"strings" //manipular cadenas de texto
)

// ER  mkdisk -size=3000 -unit=K -fit=BF -path=/home/cerezo/Disks/disk1.bin

//input := "mkdisk -size=3000 -unit=K -fit=BF -path=/home/cerezo/Disks/disk1.bin"
//mkdisk -size=200 -unit=M -path=C:/Users/Saul/Desktop/DISCO.dk
/*
parts[0] es "mkdisk"
*/
var re = regexp.MustCompile(`-(\w+)(?:=("[^"]+"|\S+))?`)

func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input

	/*Después de procesar la entrada:
	command será "mkdisk".
	params será "-size=3000 -unit=K -fit=BF -path=/home/cerezo/Disks/disk1.bin".*/
}

func Analyze() {

	for true {
		var input string
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>> INGRESE UN COMANDO <<<<<<<<<<<<<<<<<<<<<<<<<")
		fmt.Println("Ingrese comando: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		command, params := getCommandAndParams(input)

		fmt.Println("Comando: ", command, " - ", "Parametro: ", params)

		AnalyzeCommnad(command, params)

		//mkdisk -size=3000 -unit=K -fit=BF -path="/home/cerezo/Disks/disk1.bin"
	}
}

func AnalyzeScript(script string) string {
	var output bytes.Buffer

	// Separar por líneas
	lines := strings.Split(script, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignorar líneas vacías o comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		command, params := getCommandAndParams(line)
		output.WriteString(fmt.Sprintf("Comando: %s - Parametro: %s\n", command, params))

		// Redirigir la salida a un buffer en lugar de imprimir en consola
		oldOutput := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		AnalyzeCommnad(command, params) // Ejecuta el comando

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		os.Stdout = oldOutput

		// Agregar la salida capturada al resultado
		output.WriteString(buf.String())
		output.WriteString("\n")
	}

	return output.String()
}

func AnalyzeCommnad(command string, params string) {

	if strings.Contains(command, "mkdisk") {
		fn_mkdisk(params)
	} else if strings.Contains(command, "fdisk") {
		fn_fdisk(params)
	} else if strings.Contains(command, "mounted") {
		Management.Mounted()
	} else if strings.Contains(command, "mount") {
		fn_mount(params)
	} else if strings.Contains(command, "mkfs") {
		fn_mkfs(params)
	} else if strings.Contains(command, "mkfile") {
		fn_mkfile(params)
	} else if strings.Contains(command, "rmdisk") {
		fn_rmdisk(params)
	} else if strings.Contains(command, "login") {
		fn_login(params)
	} else if strings.Contains(command, "logout") {
		User.Logout()
	} else if strings.Contains(command, "mkgrp") {
		fn_mkgrp(params)
	} else if strings.Contains(command, "rmgrp") {
		fn_rmgrp(params)
	} else if strings.Contains(command, "salir") {
		fmt.Println("Saliendo del programa...")
		os.Exit(0) // Termina la ejecución del programa
	} else {
		fmt.Println("Error: Commando invalido o no encontrado")
	}

}

func fn_mkdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse flag
	fs.Parse(os.Args[1:])

	///----------------------------------------------------- extrae y asigna los valores de los parámetros
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}
	///-----------------------------------------------------
	/*
			Primera Iteración :
		    flagName es "size".
		    flagValue es "3000".
		    El switch encuentra que "size" es un flag reconocido, por lo que se ejecuta fs.Set("size", "3000").
		    Esto asigna el valor 3000 al flag size.

	*/

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// LLamamos a la funcion
	Management.Mkdisk(*size, *fit, *unit, *path)
}

func fn_rmdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta del disco a eliminar")

	// Parse flag
	fs.Parse(os.Args[1:])

	// Extraemos y asignamos los valores
	matches := re.FindAllStringSubmatch(params, -1)

	// Procesa los parámetros
	for _, match := range matches {
		flagName := match[1]                   // Captura y guarda el nombre del flag (en este caso, "path")
		flagValue := strings.ToLower(match[2]) // Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrado")
		}
	}

	// Validaciones
	if *path == "" {
		fmt.Println("Error: Path es requerido")
		return
	}

	// Llamamos a la función para eliminar el disco
	err := os.Remove(*path)
	if err != nil {
		fmt.Printf("Error: No se pudo eliminar el disco en la ruta %s: %v\n", *path, err)
		return
	}

	fmt.Printf("Disco en la ruta %s eliminado correctamente.\n", *path)
}

func fn_fdisk(input string) {
	// Definir flags
	//flag.ExitOnError hace que el programa termine si hay un error al analizar los flags.
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "k", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "", "Ajuste") // Dejar fit vacío por defecto

	// Parsear los flags
	fs.Parse(os.Args[1:])

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}
	if *name == "" {
		fmt.Println("Error: Name is required")
		return
	}

	// Si no se proporcionó un fit, usar el valor predeterminado "w"
	if *fit == "" {
		*fit = "wf"
	}

	// Validar fit (bf/wf/ff)
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return
	}

	// Llamar a la función
	Management.Fdisk(*size, *path, *name, *unit, *type_, *fit)
}

func fn_mount(params string) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2]) // Convertir todo a minúsculas
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Name son obligatorios")
		return
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	lowercaseName := strings.ToLower(*name)
	Management.Mount(*path, lowercaseName)
}

func fn_mkfs(input string) {
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "Id")
	type_ := fs.String("type", "", "Tipo")
	fs_ := fs.String("fs", "2fs", "Fs")

	// Parse the input string, not os.Args
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "id", "type", "fs":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Verifica que se hayan establecido todas las flags necesarias
	if *id == "" {
		fmt.Println("Error: id es un parámetro obligatorio.")
		return
	}

	if *type_ == "" {
		fmt.Println("Error: type es un parámetro obligatorio.")
		return
	}

	// Llamar a la función
	FileSystem.Mkfs(*id, *type_, *fs_)
}

func fn_login(input string) {
	// Definir las flags
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	// Parsearlas
	fs.Parse(os.Args[1:])

	// Match de flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}
	if *user == "" {
		fmt.Println("Error: El parámetro -user es obligatorio")
		return
	}
	if *pass == "" {
		fmt.Println("Error: El parámetro -pass es obligatorio")
		return
	}
	if *id == "" {
		fmt.Println("Error: El parámetro -id es obligatorio")
		return
	}
	User.Login(*user, *pass, *id)
}

func fn_mkgrp(params string) {
	fs := flag.NewFlagSet("mkgrp", flag.ExitOnError)
	name := fs.String("name", "", "Nombre del grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"") // Quitar comillas

		switch flagName {
		case "name":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no reconocida")
			return
		}
	}

	if *name == "" {
		fmt.Println("Error: El parámetro -name es obligatorio")
		return
	}

	User.MKGRP(*name)
}

func fn_rmgrp(params string) {
	fs := flag.NewFlagSet("mkgrp", flag.ExitOnError)
	name := fs.String("name", "", "Nombre del grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"") // Quitar comillas

		switch flagName {
		case "name":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no reconocida")
			return
		}
	}

	if *name == "" {
		fmt.Println("Error: El parámetro -name es obligatorio")
		return
	}

	User.RMGRP(*name)
}

func fn_mkfile(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkfile", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	r := fs.Bool("r", false, "Carpetas Padres")
	size := fs.Int("size", 0, "Tamanio")
	cont := fs.String("cont", "", "Contenido")

	// Parse flag
	fs.Parse(os.Args[1:])

	///----------------------------------------------------- extrae y asigna los valores de los parámetros
	matches := re.FindAllStringSubmatch(params, -1)
	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "size", "unit", "fit", "path")
		flagValue := strings.ToLower(match[2]) //trings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")
		fmt.Println("flagName: ", flagName)
		fmt.Println("flagValue: ", flagValue)
		switch flagName {
		case "size", "path", "cont":
			fs.Set(flagName, flagValue)
		case "r":
			fs.Set(flagName, "true")
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	

	
	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// LLamamos a la funcion
	Management.MkFile(*path, *size, *r, *cont)
}
