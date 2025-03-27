package Management

import (
	"MIA_Proyecto1/backend/Utilities"
	"MIA_Proyecto1/backend/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Estructura para representar una partición montada
type MountedPartition struct {
	Path   string
	Name   string
	ID     string
	Status byte // 0: no montada, 1: montada
	LoggedIn bool // true: usuario ha iniciado sesión, false: no ha iniciado sesión
}

// Mapa para almacenar las particiones montadas, organizadas por disco
var mountedPartitions = make(map[string][]MountedPartition)

// Función para imprimir las particiones montadas
func PrintMountedPartitions() {
	fmt.Println("Particiones montadas:")

	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}

	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s, Status: %c\n",
				partition.Name, partition.ID, partition.Path, partition.Status)
		}
	}
	fmt.Println("")
}

// MarkPartitionAsLoggedIn busca una partición por su ID y la marca como logueada (LoggedIn = true).
func MarkPartitionAsLoggedIn(id string) {
	// Recorre todas las particiones montadas en los discos.
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			// Si la partición coincide con el ID buscado, se marca como logueada.
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = true
				fmt.Printf("Partición con ID %s marcada como logueada.\n", id)
				return
			}
		}
	}
	// Si no se encuentra la partición, se muestra un mensaje de error.
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

func MarkPartitionAsLoggedOut(id string) {
	// Recorre todas las particiones montadas en los discos.
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			// Si la partición coincide con el ID buscado, se marca como logueada.
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = false
				fmt.Printf("Partición con ID %s marcada como logout.\n", id)
				return
			}
		}
	}
	// Si no se encuentra la partición, se muestra un mensaje de error.
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

func Mounted() {
	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas en el sistema.")
		return
	}

	fmt.Println("========================= Particiones Montadas =========================")

	// Iterar sobre todos los discos montados
	for diskPath, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskPath)
		for _, partition := range partitions {
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s\n",
				partition.Name, partition.ID, partition.Path)
		}
	}
}

// Función para obtener las particiones montadas
func GetMountedPartitions() map[string][]MountedPartition {
	return mountedPartitions
}

// ////////////////////////////////////////////////////////////////////////////
func Mkdisk(size int, fit string, unit string, path string) {
	fmt.Println("======INICIO MKDISK======")
	fmt.Printf("Size: %d\nFit: %s\nUnit: %s\nPath: %s\n", size, fit, unit, path)

	// Validaciones
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Println("Error: Fit debe ser 'bf', 'wf' o 'ff'")
		return
	}
	if size <= 0 {
		fmt.Println("Error: Size debe ser mayor a 0")
		return
	}
	if unit != "k" && unit != "m" {
		fmt.Println("Error: Las unidades válidas son 'k' o 'm'")
		return
	}

	// Crear archivo
	if err := Utilities.CreateFile(path); err != nil {
		fmt.Println("Error al crear archivo:", err)
		return
	}

	// Convertir tamaño a bytes
	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1024 * 1024
	}

	// Abrir archivo
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error al abrir archivo:", err)
		return
	}
	defer file.Close() // Asegura el cierre del archivo al salir de la función

	// Escribir ceros en un solo bloque en lugar de un bucle
	zeroBlock := make([]byte, size) // Crea un slice de bytes lleno de ceros
	if _, err := file.Write(zeroBlock); err != nil {
		fmt.Println("Error al escribir en el archivo:", err)
		return
	}

	// Crear MBR
	var newMBR Structs.MRB
	newMBR.MbrSize = int32(size)
	newMBR.Signature = rand.Int31()
	copy(newMBR.Fit[:], fit)

	// Obtener fecha actual en formato YYYY-MM-DD
	formattedDate := time.Now().Format("2006-01-02")
	copy(newMBR.CreationDate[:], formattedDate)

	// Escribir el MBR en el archivo
	if err := Utilities.WriteObject(file, newMBR, 0); err != nil {
		fmt.Println("Error al escribir el MBR:", err)
		return
	}

	// Leer el MBR para verificar que se escribió correctamente
	var tempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &tempMBR, 0); err != nil {
		fmt.Println("Error al leer el MBR:", err)
		return
	}

	// Imprimir el MBR leído
	Structs.PrintMBR(tempMBR)

	fmt.Println("======FIN MKDISK======")
}

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)

	// Validar fit (b/w/f)
	if fit != "bf" && fit != "ff" && fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	// Validar unit (b/k/m)
	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Println("Error: Unit must be 'b', 'k', or 'm'")
		return
	}

	// Ajustar el tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return
	}

	// Imprimir el objeto MBR
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	// Validaciones de las particiones
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			totalPartitions++
			usedSpace += TempMBR.Partitions[i].Size

			if TempMBR.Partitions[i].Type[0] == 'p' {
				primaryCount++
			} else if TempMBR.Partitions[i].Type[0] == 'e' {
				extendedCount++
			}
		}
	}

	// Validar que no se exceda el número máximo de particiones primarias y extendidas
	if totalPartitions >= 4 {
		fmt.Println("Error: No se pueden crear más de 4 particiones primarias o extendidas en total.")
		return
	}

	// Validar que solo haya una partición extendida
	if type_ == "e" && extendedCount > 0 {
		fmt.Println("Error: Solo se permite una partición extendida por disco.")
		return
	}

	// Validar que no se pueda crear una partición lógica sin una extendida
	if type_ == "l" && extendedCount == 0 {
		fmt.Println("Error: No se puede crear una partición lógica sin una partición extendida.")
		return
	}

	// Validar que el tamaño de la nueva partición no exceda el tamaño del disco
	if usedSpace+int32(size) > TempMBR.MbrSize {
		fmt.Println("Error: No hay suficiente espacio en el disco para crear esta partición.")
		return
	}

	// Determinar la posición de inicio de la nueva partición
	var gap int32 = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.Partitions[totalPartitions-1].Start + TempMBR.Partitions[totalPartitions-1].Size
	}

	// Encontrar una posición vacía para la nueva partición
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				// Crear partición primaria o extendida
				TempMBR.Partitions[i].Size = int32(size)
				TempMBR.Partitions[i].Start = gap
				copy(TempMBR.Partitions[i].Name[:], name)
				copy(TempMBR.Partitions[i].Fit[:], fit)
				copy(TempMBR.Partitions[i].Status[:], "0")
				copy(TempMBR.Partitions[i].Type[:], type_)
				TempMBR.Partitions[i].Correlative = int32(totalPartitions + 1)

				if type_ == "e" {
					// Inicializar el primer EBR en la partición extendida
					ebrStart := gap // El primer EBR se coloca al inicio de la partición extendida
					ebr := Structs.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					Utilities.WriteObject(file, ebr, int64(ebrStart))
				}

				break
			}
		}
	}

	// Manejar la creación de particiones lógicas dentro de una partición extendida
	if type_ == "l" {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				// Calcular la posición de inicio de la nueva partición lógica
				newEBRPos := ebr.PartStart + ebr.PartSize                    // El nuevo EBR se coloca después de la partición lógica anterior
				logicalPartitionStart := newEBRPos + int32(binary.Size(ebr)) // El inicio de la partición lógica es justo después del EBR

				// Ajustar el siguiente EBR
				ebr.PartNext = newEBRPos
				Utilities.WriteObject(file, ebr, int64(ebrPos))

				// Crear y escribir el nuevo EBR
				newEBR := Structs.EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				Utilities.WriteObject(file, newEBR, int64(newEBRPos))

				// Imprimir el nuevo EBR creado
				fmt.Println("Nuevo EBR creado:")
				Structs.PrintEBR(newEBR)
				fmt.Println("")

				// Imprimir todos los EBRs en la partición extendida
				fmt.Println("Imprimiendo todos los EBRs en la partición extendida:")
				ebrPos = TempMBR.Partitions[i].Start
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					Structs.PrintEBR(ebr)
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				break
			}
		}
		fmt.Println("")
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return
	}

	var TempMBR2 Structs.MRB
	// Leer el objeto nuevamente para verificar
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file after writing")
		return
	}

	// Imprimir el objeto MBR actualizado
	Structs.PrintMBR(TempMBR2)

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Println("======FIN FDISK======")
	fmt.Println("")

}


//////////////////////////////////////////////////////////////////////////////

// Función para montar particiones
func Mount(path string, name string) {
	file, err := Utilities.OpenFile(path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", path)
		return
	}
	defer file.Close()

	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		return
	}

	fmt.Printf("Buscando partición con nombre: '%s'\n", name)

	partitionFound := false
	var partition Structs.Partition
	var partitionIndex int

	// Convertir el nombre a bytes
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(name))

	// 🔹 Buscar en particiones primarias
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'p' && bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: Partición no encontrada o no es una partición primaria")
		return
	}
	// Verificar si la partición ya está montada
	if partition.Status[0] == '1' {
		fmt.Println("Error: La partición ya está montada")
		return
	}
	// 🔹 Verificar si la partición ya está montada en `mountedPartitions`
	diskID := generateDiskID(path)
	for _, p := range mountedPartitions[diskID] {
		if p.Name == name {
			fmt.Println("Error: La partición ya está montada en memoria")
			return
		}
	}

	// 📌 **Aquí corregimos la asignación de la letra**
	var letter byte
	if len(mountedPartitions) == 0 {
		letter = 'A' // Primer disco montado usa 'A'
	} else {
		// Si es un disco nuevo, asignamos la siguiente letra disponible
		if len(mountedPartitions[diskID]) == 0 {
			letter = getNextLetter()
		} else {
			letter = mountedPartitions[diskID][0].ID[len(mountedPartitions[diskID][0].ID)-1] // Usa la misma letra del disco
		}
	}
	
	// 🔹 Generar ID basado en carnet y número de partición
	carnet := "202202429"                   // Cambia por tu carnet real
	lastTwoDigits := carnet[len(carnet)-2:] // Últimos 2 dígitos
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, partitionIndex+1, letter)

	// Actualizar el estado de la partición a "montada" y asignar el ID generado a la partición.
	// `partition.Status[0]` se establece en '1' para indicar que la partición está montada.
	// `copy(partition.Id[:], partitionID)` asigna el ID generado a la partición.
	partition.Status[0] = '1'
	copy(partition.Id[:], partitionID)

	// Actualizamos el `TempMBR.Partitions[partitionIndex]` para reflejar los cambios en la partición.
	TempMBR.Partitions[partitionIndex] = partition

	// 🔹 Guardar en memoria
	mountedPartitions[diskID] = append(mountedPartitions[diskID], MountedPartition{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Escribir el MBR actualizado en el archivo utilizando la función `Utilities.WriteObject`.
	// Si la escritura falla, se imprime un mensaje de error.
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		return
	}
	// 🔹 Mensajes de confirmación
	fmt.Printf("Partición montada con ID: %s\n", partitionID)
	fmt.Println("MBR actualizado:")
	Structs.PrintMBR(TempMBR)
	fmt.Println("")
	PrintMountedPartitions()
}


// 🔹 Función para obtener la siguiente letra disponible
func getNextLetter() byte {
	highestLetter := 'A'
	for _, partitions := range mountedPartitions {
		for _, p := range partitions {
			letter := p.ID[len(p.ID)-1]
			if rune(letter) > highestLetter {
				highestLetter = rune(letter)
			}
		}
	}
	return byte(highestLetter + 1)
}







func generateDiskID(path string) string {
	return strings.ToLower(path)
}