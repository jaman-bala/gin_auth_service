package entities

type Role string

const (
	RoleSuperUser Role = "superuser"
	RoleAdmin     Role = "admin"
	RoleManager   Role = "manager"
	RoleUser      Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSuperUser, RoleAdmin, RoleUser, RoleManager:
		return true
	}
	return false
}

func (r Role) String() string {
	return string(r)
}

func (r Role) Level() int {
	switch r {
	case RoleSuperUser:
		return 4
	case RoleAdmin:
		return 3
	case RoleManager:
		return 2
	case RoleUser:
		return 1
	}
	return 0
}

func (r Role) HasLevel(minLevel int) bool {
	return r.Level() >= minLevel
}
