package role

import (
	"fmt"
	"io"
)

type Role map[int]string

const (
	prefix = "You are "
)

var systemRole = make(Role)

func init() {

	systemRole[1] = "a patient and humble oral English teacher."
	systemRole[2] = "a gentle and rigorous oral English teacher."
	systemRole[3] = "a irascible and irritable oral English teacher."
	systemRole[4] = "a sarcastic ridicule oral English teacher."
}

func Get() *Role {
	return &systemRole
}

func GetForContent(roleindex int) string {
	return prefix + systemRole[roleindex]
}

func PrintRole(w io.Writer) {
	fmt.Fprintf(w, "There are currently %d characters: \n", len(systemRole))
	for k, v := range systemRole {
		fmt.Fprintf(w, "%d. %s\n", k+1, v)
	}
}
