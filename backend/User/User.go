package User

import (
	"MIA_Proyecto1/backend/Management"
	"MIA_Proyecto1/backend/Structs"
	"MIA_Proyecto1/backend/Utilities"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func Login(user string, pass string, id string) {
	fmt.Println("======Start LOGIN======")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	fmt.Println("Id:", id)

	// Obtener las particiones montadas
	mountedPartitions := Management.GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var login bool = false

	// Verificar si el usuario ya está logueado en alguna partición
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id && partition.LoggedIn { // Si la partición ya tiene un usuario logueado
				fmt.Println("Error: Ya existe un usuario logueado!")
				return
			}
			if partition.ID == id { // Encuentra la partición correcta
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	// Si no se encontró la partición montada, se detiene el proceso
	if !partitionFound {
		fmt.Println("Error: No se encontró ninguna partición montada con el ID proporcionado")
		return
	}

	// Abrir el archivo del sistema de archivos binario
	file, err := Utilities.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return
	}
	defer file.Close() // Cierra el archivo al final de la ejecución

	var TempMBR Structs.MRB
	// Leer el MBR (Master Boot Record) del archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Imprimir información del MBR
	Structs.PrintMBR(TempMBR)
	fmt.Println("-------------")

	var index int = -1
	// Buscar la partición en el MBR por su ID
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 { // Verifica que la partición tenga tamaño
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) { // Compara el ID
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' { // Verifica si está montada
					fmt.Println("Partition is mounted")
					index = i
				} else {
					fmt.Println("Partition is not mounted")
					return
				}
				break
			}
		}
	}

	// Si se encontró la partición, imprimir su información
	if index != -1 {
		Structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Partition not found")
		return
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock de la partición
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return
	}

	// Buscar el archivo de usuarios "/users.txt" dentro del sistema de archivos
	indexInode := InitSearch("/users.txt", file, tempSuperblock)

	var crrInode Structs.Inode
	// Leer el Inodo del archivo "users.txt"
	if err := Utilities.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Structs.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return
	}

	// Obtener el contenido del archivo users.txt desde los bloques del inodo
	data := GetInodeFileData(crrInode, file, tempSuperblock)

	// Dividir el contenido del archivo en líneas
	lines := strings.Split(data, "\n")
	var group string
	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")

		// Si la línea tiene 5 elementos, verificar si el usuario y contraseña coinciden
		if len(words) == 5 && words[1] == "U" {
			if words[3] == user && words[4] == pass {
				login = true
				group = words[2]
				break
			}
		}
	}

	// Imprimir la información del Inodo
	fmt.Println("Inode", crrInode.I_block)

	// Si las credenciales son correctas, marcar la partición como logueada
	if login {
		fmt.Println("Usuario logueado con éxito")
		Management.MarkPartitionAsLoggedIn(id)  // Marca la partición como logueada
		StartSession(user, group, id, filepath) // Inicia la sesión
		PrintActiveSession()                    // Imprime la sesión activa
	} else {
		fmt.Println("Error: Credenciales incorrectas")
	}
	fmt.Println("======End LOGIN======")
}

func Logout() {
	fmt.Println("======Start LOGOUT======")
	if ActiveSession.IsActive {
		fmt.Println("Sesión cerrada del usuario:", ActiveSession.User)
		Management.MarkPartitionAsLoggedOut(ActiveSession.ID) // Marca la partición como deslogueada
		ActiveSession = Session{}                             // Reinicia todo
	} else {
		fmt.Println("Error: Debe de haber una sesión iniciada para hacer logout.")
	}
	fmt.Println("======End LOGOUT======")
}
func InitSearch(path string, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INICIAL ======")
	fmt.Println("path:", path)

	// Dividir la ruta en partes usando "/" como separador
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:] // Omitir el primer elemento vacío si la ruta empieza con "/"

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 Structs.Inode
	// Leer el inodo raíz (primer inodo del sistema de archivos)
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1 // Retornar -1 si hubo un error al leer
	}

	fmt.Println("======End BUSQUEDA INICIAL======")

	// Llamar a la función que busca el inodo del archivo según la ruta
	return SarchInodeByPath(StepsPath, Inode0, file, tempSuperblock)
}

// stack
func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func SarchInodeByPath(StepsPath []string, Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INODO POR PATH======")

	index := int32(0) // Contador de bloques procesados en el inodo actual

	// Extrae el primer elemento del path y elimina espacios en blanco
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	fmt.Println("========== SearchedName:", SearchedName)

	// Iterar sobre los bloques del inodo
	for _, block := range Inode.I_block {
		if block != -1 { // Si el bloque es válido (no está vacío)
			if index < 13 { // Manejo de bloques directos (0-12)
				var crrFolderBlock Structs.Folderblock

				// Leer el bloque de carpeta desde el archivo binario
				if err := Utilities.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Folderblock{})))); err != nil {
					return -1
				}

				// Buscar el archivo/directorio dentro del bloque de carpeta
				for _, folder := range crrFolderBlock.B_content {
					fmt.Println("Folder === Name:", strings.TrimRight(string(folder.B_name[:]), "\x00"), "B_inodo", folder.B_inodo)

					// Si el nombre del archivo o directorio coincide
					if strings.Contains(string(folder.B_name[:]), SearchedName) {
						fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)

						if len(StepsPath) == 0 { // Si llegamos al final de la ruta
							fmt.Println("Folder found======")
							return folder.B_inodo // Retornar índice del inodo encontrado
						} else { // Si aún hay más niveles en la ruta, seguir buscando
							fmt.Println("NextInode======")
							var NextInode Structs.Inode

							// Leer el siguiente inodo desde el archivo binario
							if err := Utilities.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Structs.Inode{})))); err != nil {
								return -1
							}

							// Llamada recursiva para seguir con la búsqueda
							return SarchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}
			} else {
				fmt.Println("Manejo de bloques indirectos no implementado") // Falta implementar acceso a bloques indirectos
			}
		}
		index++ // Incrementar índice para saber si son bloques directos o indirectos
	}

	fmt.Println("======End BUSQUEDA INODO POR PATH======")
	return 0 // Si no se encontró, retornar 0
}

func GetInodeFileData(Inode Structs.Inode, file *os.File, tempSuperblock Structs.Superblock) string {
	fmt.Println("======Start CONTENIDO DEL BLOQUE======")
	index := int32(0)
	// define content as a string
	var content string

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			//Dentro de los directos
			if index < 13 {
				var crrFileBlock Structs.Fileblock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Structs.Fileblock{})))); err != nil {
					return ""
				}

				content += string(crrFileBlock.B_content[:])

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End CONTENIDO DEL BLOQUE======")
	return content
}

// MKGRP
func MKGRP(name string) error{
	fmt.Println("======Start MKGRP======")
	fmt.Println("Group name:", name)

	// Verificar si el usuario root ya está logueado
	if !ActiveSession.IsActive{
		fmt.Println("error: No hay ninguna sesión activa. Debe hacer login primero")
		return fmt.Errorf("error: No hay ninguna sesión activa. Debe hacer login primero")
	} else if ActiveSession.User != "root" {
		fmt.Println("error: Solo el usuario root puede crear grupos")
		return fmt.Errorf("error: Solo el usuario root puede crear grupos")
	}

	// Abrir el archivo del sistema de archivos binario
	file, err := Utilities.OpenFile(ActiveSession.PartitionPath)
	if err != nil {
		fmt.Println("error: No se pudo abrir el archivo:", err)
		return fmt.Errorf("error: No se pudo abrir el archivo: %v", err)
	}
	defer file.Close() // Cierra el archivo al final de la ejecución

	var TempMBR Structs.MRB
	// Leer el MBR (Master Boot Record) del archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("error: No se pudo leer el MBR: ", err)
		return fmt.Errorf("error: No se pudo leer el MBR: %v", err)
	}

	// Imprimir información del MBR
	Structs.PrintMBR(TempMBR)
	fmt.Println("-------------")

	var index int = -1
	// Buscar la partición en el MBR por su ID
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 { // Verifica que la partición tenga tamaño
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), ActiveSession.ID) { // Compara el ID
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' { // Verifica si está montada
					fmt.Println("partition is mounted")
					index = i
				} else {
					fmt.Println("partition is not mounted")
					return fmt.Errorf("partition is not mounted")
					
				}
				break
			}
		}
	}

	// Si se encontró la partición, imprimir su información
	if index != -1 {
		Structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("partition not found")
		return fmt.Errorf("partition not found")
		
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock de la partición
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("error: No se pudo leer el Superblock: ", err)
		return fmt.Errorf("error: No se pudo leer el Superblock: %v", err)
		
	}

	// Buscar el archivo de usuarios "/users.txt" dentro del sistema de archivos
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("users.txt no encontrado")
		return fmt.Errorf("users.txt no encontrado")
		 
		
	}
	var crrInode Structs.Inode
	// Leer el Inodo del archivo "users.txt"
	if err := Utilities.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Structs.Inode{})))); err != nil {
		
		fmt.Println("error: No se pudo leer el Inodo: ", err)
		return fmt.Errorf("error: No se pudo leer el Inodo: %v", err)
		
	}

	// Obtener el contenido del archivo users.txt desde los bloques del inodo
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	cleanData := GetCleanedData(data)

	// Imprimir el tamaño de los datos limpios
	fmt.Printf("Tamaño actual de los datos limpios en users.txt: %d bytes\n", len(cleanData))

	// Dividir el contenido del archivo en líneas
	lines := strings.Split(cleanData, "\n")
	indexGroup := 1
	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")
		
		// Si la línea tiene 3 elementos, obtener el indice del ultimo grupo
		if len(words) == 3 && words[1] == "G" {
			indexGroup++
			if words[2] == name && words[0] != "0" {
				fmt.Println("error: El grupo ya existe")
				return fmt.Errorf("error: El grupo ya existe")
			}else if words[2] == name && words[0] == "0" {
				fmt.Println("error: El grupo ya existe, pero está eliminado")
				return fmt.Errorf("error: El grupo ya existe, pero está eliminado")
			}
			if words[0] == "0" {
				indexGroup--
			}
			
		}
	}
	newIndex := strconv.Itoa(indexGroup) 
	newGroup := newIndex+ ",G," + name + "\n"
	fmt.Println("New group:", newGroup)
	newData := cleanData + newGroup
	fmt.Println("Complete data:")
	fmt.Println(newData)
	
	// Guardar el contenido usando la función que maneja múltiples bloques
	if err := AppendToFileBlock(&crrInode, newData, file, tempSuperblock); err != nil {
		fmt.Println("error al escribir en users.txt:", err)
		return fmt.Errorf("error al escribir en users.txt: %v", err)
	}

	/*
	// Verificar si el nuevo contenido cabe en el bloque
	if len(newData) > 64 {
		fmt.Println("Advertencia: El contenido excede el tamaño del bloque, se truncará el contenido")
		newData = newData[:64] // Truncar el contenido a 64 bytes
	}

	// Guardar los datos actualizados de nuevo en el bloque de archivo correspondiente
	var fileBlock Structs.Fileblock
	copy(fileBlock.B_content[:], newData)

	// Escribir el bloque actualizado de vuelta en el archivo
	if err := Utilities.WriteObject(file, fileBlock, int64(tempSuperblock.S_block_start+crrInode.I_block[0]*int32(binary.Size(Structs.Fileblock{})))); err != nil {
		return fmt.Errorf("error al escribir en users.txt: %v", err)
		 
	}
	
	// Verificar el contenido del bloque después de la escritura
	fmt.Println("====== Verificación del bloque de archivo ======")
	var updatedFileBlock Structs.Fileblock
	if err := Utilities.ReadObject(file, &updatedFileBlock, int64(tempSuperblock.S_block_start+crrInode.I_block[0]*int32(binary.Size(Structs.Fileblock{})))); err != nil {
		return fmt.Errorf("error al leer el bloque actualizado: %v", err)
		 
	}
	Structs.PrintFileblock(updatedFileBlock) // Imprimir el bloque de archivo actualizado*/

	fmt.Println("Grupo creado con éxito:", newGroup)
	
	fmt.Println("======End MKGRP======")
	return nil
}

//RMGRP
func RMGRP(name string) error{
	fmt.Println("======Start RMGRP======")
	fmt.Println("Group name:", name)

	// Verificar si el usuario root ya está logueado
	if !ActiveSession.IsActive{
		fmt.Println("error: No hay ninguna sesión activa. Debe hacer login primero")
		return fmt.Errorf("error: No hay ninguna sesión activa. Debe hacer login primero")
	} else if ActiveSession.User != "root" {
		fmt.Println("error: Solo el usuario root puede crear grupos")
		return fmt.Errorf("error: Solo el usuario root puede crear grupos")
	}

	// Abrir el archivo del sistema de archivos binario
	file, err := Utilities.OpenFile(ActiveSession.PartitionPath)
	if err != nil {
		fmt.Println("error: No se pudo abrir el archivo:", err)
		return fmt.Errorf("error: No se pudo abrir el archivo: %v", err)
	}
	defer file.Close() // Cierra el archivo al final de la ejecución

	var TempMBR Structs.MRB
	// Leer el MBR (Master Boot Record) del archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("error: No se pudo leer el MBR: ", err)
		return fmt.Errorf("error: No se pudo leer el MBR: %v", err)
	}

	// Imprimir información del MBR
	Structs.PrintMBR(TempMBR)
	fmt.Println("-------------")

	var index int = -1
	// Buscar la partición en el MBR por su ID
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 { // Verifica que la partición tenga tamaño
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), ActiveSession.ID) { // Compara el ID
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' { // Verifica si está montada
					fmt.Println("partition is mounted")
					index = i
				} else {
					fmt.Println("partition is not mounted")
					return fmt.Errorf("partition is not mounted")
					
				}
				break
			}
		}
	}

	// Si se encontró la partición, imprimir su información
	if index != -1 {
		Structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("partition not found")
		return fmt.Errorf("partition not found")
		
	}

	var tempSuperblock Structs.Superblock
	// Leer el Superblock de la partición
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("error: No se pudo leer el Superblock: ", err)
		return fmt.Errorf("error: No se pudo leer el Superblock: %v", err)
	}

	// Buscar el archivo de usuarios "/users.txt" dentro del sistema de archivos
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("users.txt no encontrado")
		return fmt.Errorf("users.txt no encontrado")
	}
	var crrInode Structs.Inode
	// Leer el Inodo del archivo "users.txt"
	if err := Utilities.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Structs.Inode{})))); err != nil {
		fmt.Println("error: No se pudo leer el Inodo: ", err)
		return fmt.Errorf("error: No se pudo leer el Inodo: %v", err)
	}

	// Obtener el contenido del archivo users.txt desde los bloques del inodo
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	cleanData := GetCleanedData(data)

	// Imprimir el tamaño de los datos limpios
	fmt.Printf("Tamaño actual de los datos limpios en users.txt: %d bytes\n", len(cleanData))
	fmt.Println("Data antes de eliminar el grupo:", cleanData)
	// Dividir el contenido del archivo en líneas
	lines := strings.Split(cleanData, "\n")
	var modifiedData string
	// Iterar a través de las líneas para verificar las credenciales
	for _, line := range lines {
		words := strings.Split(line, ",")
		// Si la línea tiene 3 elementos, obtener el indice del ultimo grupo
		if len(words) == 3 && words[1] == "G" {
			if words[2] == name && words[0] != "0"{
				words[0] = "0"
				modifiedData += "0,G,"+ name + "\n"
				continue
			}else if words[2] == name && words[0] == "0"{
				fmt.Println("error: El grupo ya está eliminado")
				return fmt.Errorf("error: El grupo ya está eliminado")
			}else{
				modifiedData += line+"\n"
				continue
			}
			
		}
		modifiedData += line + "\n"
	}
	cleanData = GetCleanedData(modifiedData)
	fmt.Println("Data despues de eliminar el grupo:", cleanData)
	
	
	// Guardar el contenido usando la función que maneja múltiples bloques
	if err := AppendToFileBlock(&crrInode, cleanData, file, tempSuperblock); err != nil {
		return fmt.Errorf("error al escribir en users.txt: %v", err)
	}


	fmt.Println("Grupo eliminado con éxito:", name)
	
	fmt.Println("======End RMGRP======")
	return nil
}

// MKUSER
func GetCleanedData(data string) string {
	cleanedData := strings.TrimRight(data, "\x00") // Eliminar bytes nulos del final
	return cleanedData
}

func AppendToFileBlock(inode *Structs.Inode, newData string, file *os.File, superblock Structs.Superblock) error {
	
	// Convertir a bytes
	dataBytes := []byte(newData)
	blockSize := binary.Size(Structs.Fileblock{})
	totalBlocks := int(math.Ceil(float64(len(dataBytes)) / float64(blockSize)))

	if totalBlocks > len(inode.I_block) {
		return fmt.Errorf("el archivo excede la capacidad de bloques del inodo")
	}

	// Escribir cada bloque por separado
	for i := 0; i < totalBlocks; i++ {
		start := i * blockSize
		end := start + blockSize
		if end > len(dataBytes) {
			end = len(dataBytes)
		}
		chunk := dataBytes[start:end]

		// Preparar bloque de archivo
		var block Structs.Fileblock
		copy(block.B_content[:], chunk)

		// Calcular posición en el disco
		blockIndex := inode.I_block[i]
		if blockIndex == -1 {
			blockIndex = int32(i)
			inode.I_block[i] = blockIndex // Asignar el bloque si estaba libre
		}
		offset := int64(superblock.S_block_start + blockIndex*int32(blockSize))

		// Escribir el bloque en el archivo
		if err := Utilities.WriteObject(file, block, offset); err != nil {
			return fmt.Errorf("error al escribir el bloque %d: %v", i, err)
		}
	}

	// Actualizar tamaño del inodo y guardar
	inode.I_size = int32(len(dataBytes))
	inodeOffset := int64(superblock.S_inode_start + inode.I_block[0]*int32(binary.Size(Structs.Inode{})))
	if err := Utilities.WriteObject(file, *inode, inodeOffset); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}
	
	return nil
}
