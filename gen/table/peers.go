//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/sqlite"
)

var Peers = newPeersTable("", "peers", "")

type peersTable struct {
	sqlite.Table

	//Columns
	ID       sqlite.ColumnString
	AddrInfo sqlite.ColumnString

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type PeersTable struct {
	peersTable

	EXCLUDED peersTable
}

// AS creates new PeersTable with assigned alias
func (a PeersTable) AS(alias string) *PeersTable {
	return newPeersTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PeersTable with assigned schema name
func (a PeersTable) FromSchema(schemaName string) *PeersTable {
	return newPeersTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new PeersTable with assigned table prefix
func (a PeersTable) WithPrefix(prefix string) *PeersTable {
	return newPeersTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new PeersTable with assigned table suffix
func (a PeersTable) WithSuffix(suffix string) *PeersTable {
	return newPeersTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newPeersTable(schemaName, tableName, alias string) *PeersTable {
	return &PeersTable{
		peersTable: newPeersTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newPeersTableImpl("", "excluded", ""),
	}
}

func newPeersTableImpl(schemaName, tableName, alias string) peersTable {
	var (
		IDColumn       = sqlite.StringColumn("id")
		AddrInfoColumn = sqlite.StringColumn("addr_info")
		allColumns     = sqlite.ColumnList{IDColumn, AddrInfoColumn}
		mutableColumns = sqlite.ColumnList{AddrInfoColumn}
	)

	return peersTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:       IDColumn,
		AddrInfo: AddrInfoColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
