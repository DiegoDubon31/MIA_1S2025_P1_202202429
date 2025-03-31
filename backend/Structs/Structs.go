package Structs

import (
	"fmt"
	"strings"
)

type MRB struct {
	MbrSize      int32
	CreationDate [10]byte
	Signature    int32
	Fit          [2]byte
	Partitions   [4]Partition
}

func PrintMBR(data MRB) {
	fmt.Printf("CreationDate: %s, fit: %s, size: %d\n", strings.TrimRight(string(data.CreationDate[:]), "\x00"), strings.TrimRight(string(data.Fit[:]), "\x00"), data.MbrSize)
	for i := 0; i < 4; i++ {
		PrintPartition(data.Partitions[i])
	}
}

type Partition struct {
	Status      [1]byte
	Type        [1]byte
	Fit         [1]byte
	Start       int32
	Size        int32
	Name        [16]byte
	Correlative int32
	Id          [4]byte
}

func PrintPartition(data Partition) {
	fmt.Printf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s\n",
		strings.TrimRight(string(data.Name[:]), "\x00"),
		strings.TrimRight(string(data.Type[:]), "\x00"),
		data.Start,
		data.Size,
		strings.TrimRight(string(data.Status[:]), "\x00"),
		strings.TrimRight(string(data.Id[:]), "\x00"))
}

type EBR struct {
	PartMount byte
	PartFit   byte
	PartStart int32
	PartSize  int32
	PartNext  int32
	PartName  [16]byte
}

func PrintEBR(data EBR) {
	fmt.Printf("Name: %s, fit: %c, start: %d, size: %d, next: %d, mount: %c\n",
		strings.TrimRight(string(data.PartName[:]), "\x00"),
		data.PartFit,
		data.PartStart,
		data.PartSize,
		data.PartNext,
		data.PartMount)
}

type Superblock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             [17]byte
	S_umtime            [17]byte
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_fist_ino          int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

func PrintSuperblock(sb Superblock) {
	fmt.Println("====== Superblock ======")
	fmt.Printf("S_filesystem_type: %d\n", sb.S_filesystem_type)
	fmt.Printf("S_inodes_count: %d\n", sb.S_inodes_count)
	fmt.Printf("S_blocks_count: %d\n", sb.S_blocks_count)
	fmt.Printf("S_free_blocks_count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("S_free_inodes_count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("S_mtime: %s\n", strings.TrimRight(string(sb.S_mtime[:]), "\x00"))
	fmt.Printf("S_umtime: %s\n", strings.TrimRight(string(sb.S_umtime[:]), "\x00"))
	fmt.Printf("S_mnt_count: %d\n", sb.S_mnt_count)
	fmt.Printf("S_magic: 0x%X\n", sb.S_magic)
	fmt.Printf("S_inode_size: %d\n", sb.S_inode_size)
	fmt.Printf("S_block_size: %d\n", sb.S_block_size)
	fmt.Printf("S_fist_ino: %d\n", sb.S_fist_ino)
	fmt.Printf("S_first_blo: %d\n", sb.S_first_blo)
	fmt.Printf("S_bm_inode_start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("S_bm_block_start: %d\n", sb.S_bm_block_start)
	fmt.Printf("S_inode_start: %d\n", sb.S_inode_start)
	fmt.Printf("S_block_start: %d\n", sb.S_block_start)
	fmt.Println("========================")
}

type Inode struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime [17]byte
	I_ctime [17]byte
	I_mtime [17]byte
	I_block [15]int32
	I_type  [1]byte
	I_perm  [3]byte
}

func PrintInode(inode Inode) {
	fmt.Println("====== Inode ======")
	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", strings.TrimRight(string(inode.I_atime[:]), "\x00"))
	fmt.Printf("I_ctime: %s\n", strings.TrimRight(string(inode.I_ctime[:]), "\x00"))
	fmt.Printf("I_mtime: %s\n", strings.TrimRight(string(inode.I_mtime[:]), "\x00"))
	fmt.Printf("I_type: %s\n", strings.TrimRight(string(inode.I_type[:]), "\x00"))
	fmt.Printf("I_perm: %s\n", strings.TrimRight(string(inode.I_perm[:]), "\x00"))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Println("===================")
}

type Folderblock struct {
	B_content [4]Content
}

func PrintFolderblock(folderblock Folderblock) {
	fmt.Println("====== Folderblock ======")
	for i, content := range folderblock.B_content {
		fmt.Printf("Content %d: Name: %s, Inodo: %d\n", i, strings.TrimRight(string(content.B_name[:]), "\x00"), content.B_inodo)
	}
	fmt.Println("=========================")
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type Fileblock struct {
	B_content [64]byte
}

func PrintFileblock(fileblock Fileblock) {
	fmt.Println("====== Fileblock ======")
	fmt.Printf("B_content: %s\n", strings.TrimRight(string(fileblock.B_content[:]), "\x00"))
	fmt.Println("=======================")
}

type Pointerblock struct {
	B_pointers [16]int32
}

func PrintPointerblock(pointerblock Pointerblock) {
	fmt.Println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_pointers {
		fmt.Printf("Pointer %d: %d\n", i, pointer)
	}
	fmt.Println("=========================")
}
