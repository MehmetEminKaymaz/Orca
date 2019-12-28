package Orca

import (
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

const(//supported databases
	MySQL="MySQL"
	SQLite="SQLite"
	MSSQL="MSSQL"
	MongoDB="MongoDB"
	Redis="Redis"
	PostgreSQL="PostgreSQL"
	MemCached="MemCached"
)

type IDatabase interface {

	GetCollection(interface{},string) ICollection


}

type IOptions interface {
	Options() []string
}

type ICollection interface {

	Add(interface{})
	AddRange(interface{})
	Update(interface{},interface{})
	Delete(interface{})
	Clear()
	Foreach(interface{})
	GetLogs()
	ToSlice() []interface{}

}


func Use(yourDb string,options IOptions) (IDatabase,error){

	switch yourDb {

	case MySQL:
		return nil,nil
	case MSSQL:
		return nil,nil
	case SQLite:
		dname:=options.Options()[0]
		path:=options.Options()[1]
		return getDatabase(dname,path),nil
	case PostgreSQL:
		return nil,nil
	case MongoDB:
		applyUri:=options.Options()[0]
		dname:=options.Options()[1]
		return getDatabaseMongo(applyUri,dname),nil
	case Redis:
		addr:=options.Options()[0]
		pass:=options.Options()[1]
		db:=options.Options()[2]
		return getDatabaseRedis(addr,pass,db),nil
	case MemCached:
		s:=options.Options()
		return getDatabaseMemCached(s),nil
	default:
		//we do not support
		return nil,errors.New("Orca does not support the database.(Supported Databases : MySQL,SQLite," +
			"MSSQL,MongoDB,Redis,PostgreSQL)")

	}
}




















