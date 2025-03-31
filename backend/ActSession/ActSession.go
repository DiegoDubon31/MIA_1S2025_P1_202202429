package ActSession

import "fmt"

type Session struct {
	User          string
	Group         string
	ID            string
	PartitionPath string
	IsActive      bool
}

var ActiveSession Session

func CheckLogin() bool {
	if !ActiveSession.IsActive {
		fmt.Println("Error: No hay ninguna sesión activa. Debe hacer login primero.")
		return false
	}
	return true
}


func StartSession(user string, group string, id string, path string) {
	ActiveSession = Session{
		User:     user,
		Group:    group,
		ID:       id,
		PartitionPath:     path,
		IsActive: true,
	}
}

func GetSession() Session {
	return ActiveSession
}
func PrintActiveSession() {
	fmt.Println("======Active Session======")
	fmt.Println("User:", ActiveSession.User)
	fmt.Println("Group:", ActiveSession.Group)
	fmt.Println("ID:", ActiveSession.ID)
	fmt.Println("Partition Path:", ActiveSession.PartitionPath)
	fmt.Println("Active:", ActiveSession.IsActive)
	fmt.Println("===========================")
}
