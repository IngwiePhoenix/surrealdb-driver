package gorm

type RootInfo struct {
	Accesses   map[string]string `gorm:"column:accesses"`
	Namespaces map[string]string `gorm:"column:namespaces"`
	Nodes      map[string]string `gorm:"column:nodes"`
	Users      map[string]string `gorm:"column:users"`
}

type NamespaceInfo struct {
	Accesses  map[string]string `gorm:"column:accesses"`
	Databases map[string]string `gorm:"column:databases"`
	Users     map[string]string `gorm:"column:users"`
}

type DatabaseInfo struct {
	Accesses  map[string]string `gorm:"column:accesses"`
	Configs   map[string]string `gorm:"column:configs"`
	Functions map[string]string `gorm:"column:functions"`
	Models    map[string]string `gorm:"column:models"`
	Params    map[string]string `gorm:"column:params"`
	Tables    map[string]string `gorm:"column:tables"`
	Users     map[string]string `gorm:"column:users"`
}

type TableInfo struct {
	Events  map[string]string `gorm:"column:events"`
	Fields  map[string]string `gorm:"column:fields"`
	Indexes map[string]string `gorm:"column:indexes"`
	Lives   map[string]string `gorm:"column:lives"`
	Tables  map[string]string `gorm:"column:tables"`
}

// This returns just a string.
// Probably a better idea to just type-alias it...?
// type UserInfo struct{}
type UserInfo = string

// Nested objects are kinda difficult.
// That said, GORM _should_ be able to recognize this.
type IndexInfo struct {
	Building struct {
		Count  *int   `gorm:"column:count"`
		Status string `gorm:"column:status"`
	} `gorm:"column:building"`
}
