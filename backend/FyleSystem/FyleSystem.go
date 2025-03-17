package FileSystem

import (
	"backend/services"
	"backend/structs"
	"backend/models"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Mkfs(id string, type_ string, fs_ string) {
	fmt.Println("======INICIO MKFS======")
	fmt.Println("Id:", id)
	fmt.Println("Type:", type_)
	fmt.Println("Fs:", fs_)

	// Buscar la partición montada por ID
	var mountedPartition models.MountedPartition
	var partitionFound bool

	// Iterar sobre las particiones montadas y buscar la partición que coincida con el ID
	for _, partitions := range models.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	// Si no se encuentra la partición, se sale de la función
	if !partitionFound {
		fmt.Println("Particion no encontrada")
		return
	}

	// Verificar si la partición está montada (estado '1')
	if mountedPartition.Status != '1' {
		fmt.Println("La particion aun no esta montada")
		return
	}

	// Abrir el archivo binario de la partición
	file, err := services.OpenFile(mountedPartition.Path)
	if err != nil {
		return
	}

	// Leer el MBR (Master Boot Record) del archivo binario
	var TempMBR structs.MRB
	if err := services.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	// Imprimir el MBR para depuración
	structs.PrintMBR(TempMBR)

	// Buscar la partición dentro del MBR que coincida con el ID proporcionado
	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				index = i
				break
			}
		}
	}

	// Si no se encuentra la partición, se sale de la función
	if index != -1 {
		structs.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Particion no encontrada (2)")
		return
	}

	// Calcular el número de inodos basado en el tamaño de la partición
	numerador := int32(TempMBR.Partitions[index].Size - int32(binary.Size(structs.Superblock{})))
	denominador_base := int32(4 + int32(binary.Size(structs.Inode{})) + 3*int32(binary.Size(structs.Fileblock{})))
	var temp int32 = 0
	if fs_ == "2fs" {
		temp = 0
	} else {
		fmt.Print("Error por el momento solo está disponible 2FS.")
	}
	denominador := denominador_base + temp
	n := int32(numerador / denominador)

	fmt.Println("INODOS:", n)

	// Crear el Superblock con los campos calculados
	var newSuperblock structs.Superblock
	newSuperblock.S_filesystem_type = 2 // EXT2
	newSuperblock.S_inodes_count = n
	newSuperblock.S_blocks_count = 3 * n
	newSuperblock.S_free_blocks_count = 3*n - 2
	newSuperblock.S_free_inodes_count = n - 2
	copy(newSuperblock.S_mtime[:], "25/02/2025")
	copy(newSuperblock.S_umtime[:], "25/02/2025")
	newSuperblock.S_mnt_count = 1
	newSuperblock.S_magic = 0xEF53
	newSuperblock.S_inode_size = int32(binary.Size(structs.Inode{}))
	newSuperblock.S_block_size = int32(binary.Size(structs.Fileblock{}))

	// Calcular las posiciones de inicio de los bloques en el disco
	newSuperblock.S_bm_inode_start = TempMBR.Partitions[index].Start + int32(binary.Size(structs.Superblock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + n
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*n
	newSuperblock.S_block_start = newSuperblock.S_inode_start + n*newSuperblock.S_inode_size

	// Si el sistema de archivos es 2FS, se crea un sistema de archivos EXT2
	if fs_ == "2fs" {
		create_ext2(n, TempMBR.Partitions[index], newSuperblock, "25/02/2025", file)
	} else {
		fmt.Println("EXT3 no está soportado.")
	}

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Println("======FIN MKFS======")
}

// Función auxiliar para crear el sistema de archivos EXT2
func create_ext2(n int32, partition structs.Partition, newSuperblock structs.Superblock, date string, file *os.File) {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("INODOS:", n)

	// Imprimir el Superblock calculado
	structs.PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	// Escribir los bitmaps de inodos y bloques
	for i := int32(0); i < n; i++ {
		if err := services.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	// Escribir los bitmaps de bloques
	for i := int32(0); i < 3*n; i++ {
		if err := services.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	// Inicializar los inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Crear la carpeta raíz y el archivo "users.txt"
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Escribir el superbloque actualizado en el archivo
	if err := services.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Marcar los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Leer e imprimir los inodos después de formatear
	fmt.Println("====== Imprimiendo Inodos ======")
	for i := int32(0); i < n; i++ {
		var inode structs.Inode
		offset := int64(newSuperblock.S_inode_start + i*int32(binary.Size(structs.Inode{})))
		if err := services.ReadObject(file, &inode, offset); err != nil {
			fmt.Println("Error al leer inodo: ", err)
			return
		}
		structs.PrintInode(inode)
	}

	// Leer e imprimir los Folderblocks y Fileblocks
	fmt.Println("====== Imprimiendo Folderblocks y Fileblocks ======")

	// Imprimir Folderblocks
	for i := int32(0); i < 1; i++ {
		var folderblock structs.Folderblock
		offset := int64(newSuperblock.S_block_start + i*int32(binary.Size(structs.Folderblock{})))
		if err := services.ReadObject(file, &folderblock, offset); err != nil {
			fmt.Println("Error al leer Folderblock: ", err)
			return
		}
		structs.PrintFolderblock(folderblock)
	}

	// Imprimir Fileblocks
	for i := int32(0); i < 1; i++ {
		var fileblock structs.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(structs.Folderblock{})) + i*int32(binary.Size(structs.Fileblock{})))
		if err := services.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return
		}
		structs.PrintFileblock(fileblock)
	}

	// Imprimir el Superblock final
	structs.PrintSuperblock(newSuperblock)

	fmt.Println("======End CREATE EXT2======")
}

// Función auxiliar para inicializar inodos y bloques
func initInodesAndBlocks(n int32, newSuperblock structs.Superblock, file *os.File) error {
	var newInode structs.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := services.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(structs.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock structs.Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := services.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(structs.Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock structs.Superblock, date string, file *os.File) error {
	var Inode0, Inode1 structs.Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	// Asignar el tamaño real del contenido
	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize // Esto ahora refleja el tamaño real del contenido

	var Fileblock1 structs.Fileblock
	copy(Fileblock1.B_content[:], data) // Copia segura de datos a Fileblock

	var Folderblock0 structs.Folderblock
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	// Escribir los inodos y bloques en las posiciones correctas
	if err := services.WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := services.WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(structs.Inode{})))); err != nil {
		return err
	}
	if err := services.WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := services.WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(structs.Folderblock{})))); err != nil {
		return err
	}

	return nil
}

// Función auxiliar para inicializar un inodo
func initInode(inode *structs.Inode, date string) {
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	copy(inode.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_block[i] = -1
	}
}

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock structs.Superblock, file *os.File) error {
	if err := services.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := services.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := services.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := services.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}