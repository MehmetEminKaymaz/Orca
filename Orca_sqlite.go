package Orca
/*

 SQLite

*/
import (
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Database struct {

	Name string
	db *sql.DB
	AllTables []string
	LocalH []LocalHook
	lock bool


}

type Table struct {
	Name string
	Columns []string
	ColTypes []reflect.Type
	ContainsEmbedded bool
	EmbeddedTables []Table
	ContainsArr bool
	ArrTables []Table
}

type Collection struct {
	ITsTable Table
	List map[int]interface{}
	data *Database
	LocalH []LocalHook
	lock bool

}

type SQLiteOptions struct {
    driverName string
    path string
}

func SetSQLiteOptions(driverName,path string) IOptions{
	return SQLiteOptions{
		path:path,
		driverName:driverName,
	}
}

func(options SQLiteOptions) Options() []string{
  var ops []string

  ops=append(ops,options.driverName)
  ops=append(ops,options.path)

  return ops
}



func getDatabase(driverName,path string) *Database{

	db,err:=sql.Open(driverName,path)
	Check(err)

	return &Database{
		Name:filepath.Base(path),
		db:db,
		LocalH:[]LocalHook{},
		lock:false,
	}
}

//will be implement immediately
func(Data *Database) AddLocalHooks(hks ...LocalHook){

	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	Data.DeleteLocalHooks(ids...)
	Data.LocalH=append(Data.LocalH,hks...)

}
func(Data *Database) AddLocalHook(hks LocalHook){

	Data.DeleteLocalHook(hks.getID())
	Data.LocalH=append(Data.LocalH,hks)

}
func(Data *Database) DeleteLocalHook(hks string){
	for i:=0;i< len(Data.LocalH);i++{
		if Data.LocalH[i].getID()==hks{
			Data.LocalH[i]=Data.LocalH[len(Data.LocalH)-1]
			Data.LocalH=Data.LocalH[:len(Data.LocalH)-1]
			break
		}
	}
}
func(Data *Database) DeleteLocalHooks(hks ...string){
	Data.LocalH=reorder(Data.LocalH,hks)
}
//




func(Data *Database)  GetCollection(x interface{},tableName string) ICollection {

	var t Table = getTable(x,tableName) //get table from struct(x's underlying value must be a struct!)


	execTheQueries(Data,getQueries(t))//we create tables if not exist

	getAllTableNamesforDb(Data)




	myMap:=make(map[int]interface{})
	//now we get tuples from db and convert to our struct type(to x's underlying value)



	count:=getTupleCount(Data,tableName)

	if count<=0{//table is empty


	}else{

		min:=getMinId(Data,tableName)
		max:=getMAXId(Data,tableName)

		for i:=min;i<=max;i++{
			s:=strconv.Itoa(i)
			myMap[i]=getTupleToStruct(Data,x,s,tableName).Interface()
		}

	}





	//
	_=t

	return &Collection{
		ITsTable:t,
		List:myMap,
		data:Data,
		LocalH:Data.LocalH,
		lock:false,

	}





}



func getMAXId(Data *Database,tableName string) int{
	var result int
	row:=Data.db.QueryRow("SELECT MAX(Id) FROM "+tableName)
	err:=row.Scan(&result)
	Check(err)
	return result
}



func getMinId(Data *Database,tableName string) int{
	var result int
	row:=Data.db.QueryRow("SELECT MIN(Id) FROM "+tableName)
	err:=row.Scan(&result)
	Check(err)

	return result

}


func getTupleToStruct2(Data *Database,x interface{},Id,tableName string) interface{} {
	//count := getTupleCount(Data,tableName)

	//ptr:=reflect.New(reflect.TypeOf(x))
	//newStruct:=ptr.Elem()
	//fmt.Println(newStruct.Type())


	NewStruct:=reflect.ValueOf(&x).Elem()
	fmt.Println(NewStruct.CanAddr())

	BossSlice:=NewSliceFor(NewStruct,0,1)
	ColNames:=getColumnNames(Data, tableName)
	ColNums:= len(ColNames)

	cachePS:=make([]interface{},ColNums,ColNums)
	cache :=make([]interface{},ColNums,ColNums)
	for i,_:=range cachePS{
		cache[i]=&cachePS[i]
	}

	rows,err:=Data.db.Query("SELECT * FROM "+tableName+" WHERE Id=="+Id)
	Check(err)
	for rows.Next(){
		err:=rows.Scan(cache...)
		Check(err)

		for i,v:=range ColNames[1:]{//[1:] jump over Id column!

			if strings.Contains(v,"EMARR_"){//embedded array or slice!

				s:=strings.Split(v,"_")

				switch s[1] {
				case "int"://caches array of int
					TheSlice:=make([]int,0,1)
					var cacheId int
					var cacheInt int
					rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
					Check(err)
					for rows.Next(){
						err=rows.Scan(&cacheId,&cacheInt)
						Check(err)
						TheSlice=append(TheSlice,cacheInt)
					}
					defer func(){
						err:=rows.Close()
						Check(err)
					}()

					NewStruct.Field(i).Set(reflect.ValueOf(TheSlice))



				case "string": //caches array of string

					TheSlice:=make([]string,0,1)
					var cacheId int
					var cacheString string
					rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
					Check(err)
					for rows.Next(){
						err=rows.Scan(&cacheId,&cacheString)
						Check(err)
						TheSlice=append(TheSlice,cacheString)
					}
					defer func(){
						err:=rows.Close()
						Check(err)
					}()
					NewStruct.Field(i).Set(reflect.ValueOf(TheSlice))


				case "float": //caches array of float
					TheSlice:=make([]float32,0,1)
					var cacheId int
					var cachefloat float32
					rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
					Check(err)
					for rows.Next(){
						err=rows.Scan(&cacheId,&cachefloat)
						Check(err)
						TheSlice=append(TheSlice,cachefloat)
					}
					defer func(){
						err:=rows.Close()
						Check(err)
					}()

					NewStruct.Field(i).Set(reflect.ValueOf(TheSlice))

				default://default catches array of struct


					/*	val:=getTupleToStruct2(Data,newStruct.Field(i).Interface(),Id,tableName+"_"+s[3])
						reflect.ValueOf(newStruct.Field(i))=reflect.Append(reflect.ValueOf(newStruct.Field(i)),val...)
		*/
					/*val:=getTupleToStruct2(Data,newStruct.Field(i).Type().Elem().Name(),Id,tableName+"_"+s[2])
					newStruct.Field(i).Set(reflect.ValueOf(val))*/
					val:=getTupleToStruct2(Data,NewStruct.Field(i).Type().Elem(),Id,tableName+"_"+s[2])
					NewStruct.Field(i).Elem().Set(reflect.ValueOf(val))

				}


			}else if strings.Contains(v,"EM_"){//embedded struct!

				s:=strings.Split(v,"_")
				EMId:=int(cachePS[i+1].(int64))
				sEMId:=strconv.Itoa(EMId)
				TheOtherX:=NewStruct.Field(i).Interface()
				NewStruct.Field(i).Set(getTupleToStruct(Data,TheOtherX,sEMId,tableName+"_"+s[1]))



			}else if v=="Id"{//primary key or foreign key

				//reflect.ValueOf(newStruct).Field(i).Set(reflect.ValueOf(int(slice[i].(int64))))


			}else{//normal

				s:=strings.Split(v,"_")
				switch s[1] {
				case "int":
					val:=reflect.ValueOf(int(cachePS[i+1].(int64)))
					_=val
					NewStruct.Field(i).Set(val)
				case "string":
					val:=reflect.ValueOf(cachePS[i+1].(string))
					//fmt.Println(reflect.ValueOf(&NewStruct).Elem().CanAddr())
					//fmt.Println(reflect.ValueOf(&NewStruct).Elem().Field(i).CanAddr())
					NewStruct.Field(i).Set(val)

				case "float":
					val:=reflect.ValueOf(float32(cachePS[i+1].(float64)))
					NewStruct.Field(i).Set(val)

				default:


				}


			}

		}

		BossSlice=reflect.Append(reflect.ValueOf(BossSlice),NewStruct)
	}
	defer func(){
		err:=rows.Close()
		Check(err)
	}()



	return BossSlice
}


func getTupleToStructForSlices(Data *Database,x interface{},Id,tableName string) reflect.Value{



	//newStruct:=reflect.ValueOf(newStructLike(x))
	//newStruct:=reflect.ValueOf(x)
	ptr:=reflect.New(reflect.TypeOf(x))
	newStruct:=ptr.Elem()

	ColNames:=getColumnNames(Data, tableName)
	ColNums:= len(ColNames)


	cachePS:=make([]interface{},ColNums,ColNums)
	cache :=make([]interface{},ColNums,ColNums)
	for i,_:=range cachePS{
		cache[i]=&cachePS[i]
	}

	row:=Data.db.QueryRow("SELECT * FROM "+tableName+" WHERE rowid=="+Id)


	err:=row.Scan(cache...)

	Check(err)

	for i,v:=range ColNames[1:]{//[1:] jump over Id column!

		if strings.Contains(v,"EMARR_"){//embedded array or slice!

			s:=strings.Split(v,"_")

			switch s[1] {
			case "int"://caches array of int
				TheSlice:=make([]int,0,1)
				var cacheId int
				var cacheInt int
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cacheInt)
					Check(err)
					TheSlice=append(TheSlice,cacheInt)
				}

				err=rows.Close()
				Check(err)


				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))



			case "string": //caches array of string

				TheSlice:=make([]string,0,1)
				var cacheId int
				var cacheString string
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cacheString)
					Check(err)
					TheSlice=append(TheSlice,cacheString)
				}

				err=rows.Close()
				Check(err)

				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))


			case "float": //caches array of float
				TheSlice:=make([]float32,0,1)
				var cacheId int
				var cachefloat float32
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cachefloat)
					Check(err)
					TheSlice=append(TheSlice,cachefloat)
				}

				err=rows.Close()
				Check(err)

				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))

			default://default catches array of struct

				//TheSlice:=make([]interface{},0,1)

				cacheStruct:=newStruct.Field(i)



				elemType := cacheStruct.Type().Elem()

				elemSlice := reflect.MakeSlice(reflect.SliceOf(elemType), 0,0)

				min:=getMinId(Data,tableName+"_"+s[2])
				max:=getMAXId(Data,tableName+"_"+s[2])
				_=max
				_=min






						fmt.Println("Id:" + strconv.Itoa(max))

						intPtr := reflect.New(cacheStruct.Type().Elem())
						b := intPtr.Elem().Interface()

						strc := getTupleToStruct(Data, b, Id, tableName+"_"+s[2])

						elemSlice = reflect.Append(elemSlice, strc)
						//TheSlice=append(TheSlice,strc)




				newStruct.Field(i).Set(elemSlice)



			}


		}else if strings.Contains(v,"EM_"){//embedded struct!

			s:=strings.Split(v,"_")
			EMId:=int(cachePS[i+1].(int64))
			sEMId:=strconv.Itoa(EMId)
			TheOtherX:=newStruct.Field(i).Interface()
			newStruct.Field(i).Set(getTupleToStruct(Data,TheOtherX,sEMId,tableName+"_"+s[1]))



		}else if v=="Id"{//primary key or foreign key

			//reflect.ValueOf(newStruct).Field(i).Set(reflect.ValueOf(int(slice[i].(int64))))


		}else{//normal

			s:=strings.Split(v,"_")
			switch s[1] {
			case "int":
				val:=reflect.ValueOf(int(cachePS[i+1].(int64)))
				_=val
				newStruct.Field(i).Set(val)
			case "string":
				val:=reflect.ValueOf(cachePS[i+1].(string))
				newStruct.Field(i).Set(val)

			case "float":
				val:=reflect.ValueOf(float32(cachePS[i+1].(float64)))
				newStruct.Field(i).Set(val)

			default:


			}


		}

	}



	return newStruct
}





func getTupleToStruct(Data *Database,x interface{},Id,tableName string) reflect.Value {//x must be a struct



	//newStruct:=reflect.ValueOf(newStructLike(x))
	//newStruct:=reflect.ValueOf(x)
	ptr:=reflect.New(reflect.TypeOf(x))
	newStruct:=ptr.Elem()

	ColNames:=getColumnNames(Data, tableName)
	ColNums:= len(ColNames)


	cachePS:=make([]interface{},ColNums,ColNums)
	cache :=make([]interface{},ColNums,ColNums)
	for i,_:=range cachePS{
		cache[i]=&cachePS[i]
	}

	row:=Data.db.QueryRow("SELECT * FROM "+tableName+" WHERE Id=="+Id)


	err:=row.Scan(cache...)

	Check(err)

	for i,v:=range ColNames[1:]{//[1:] jump over Id column!

		if strings.Contains(v,"EMARR_"){//embedded array or slice!

			s:=strings.Split(v,"_")

			switch s[1] {
			case "int"://caches array of int
				TheSlice:=make([]int,0,1)
				var cacheId int
				var cacheInt int
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cacheInt)
					Check(err)
					TheSlice=append(TheSlice,cacheInt)
				}

				err=rows.Close()
				Check(err)


				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))



			case "string": //caches array of string

				TheSlice:=make([]string,0,1)
				var cacheId int
				var cacheString string
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cacheString)
					Check(err)
					TheSlice=append(TheSlice,cacheString)
				}

				err=rows.Close()
				Check(err)

				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))


			case "float": //caches array of float
				TheSlice:=make([]float32,0,1)
				var cacheId int
				var cachefloat float32
				rows,err:=Data.db.Query("SELECT * FROM "+tableName+"_"+s[2]+" WHERE Id=="+Id)
				Check(err)
				for rows.Next(){
					err=rows.Scan(&cacheId,&cachefloat)
					Check(err)
					TheSlice=append(TheSlice,cachefloat)
				}

				err=rows.Close()
				Check(err)

				newStruct.Field(i).Set(reflect.ValueOf(TheSlice))

			default://default catches array of struct

				//TheSlice:=make([]interface{},0,1)

				cacheStruct:=newStruct.Field(i)



				elemType := cacheStruct.Type().Elem()

				elemSlice := reflect.MakeSlice(reflect.SliceOf(elemType), 0,0)

				min:=getMinId(Data,tableName+"_"+s[2])
				max:=getMAXId(Data,tableName+"_"+s[2])
				_=max
				for i:=min;i<=max;i++{

					tupleCount:=getTupleCountWithId(Data,tableName+"_"+s[2],i)

					for k:=1;k<=tupleCount;k++ {

						fmt.Println("Id:" + strconv.Itoa(max))

						intPtr := reflect.New(cacheStruct.Type().Elem())
						b := intPtr.Elem().Interface()

						strc := getTupleToStructForSlices(Data, b, strconv.Itoa(k), tableName+"_"+s[2])

						elemSlice = reflect.Append(elemSlice, strc)
						//TheSlice=append(TheSlice,strc)
					}

				}

				newStruct.Field(i).Set(elemSlice)



			}


		}else if strings.Contains(v,"EM_"){//embedded struct!

			s:=strings.Split(v,"_")
			EMId:=int(cachePS[i+1].(int64))
			sEMId:=strconv.Itoa(EMId)
			TheOtherX:=newStruct.Field(i).Interface()
			newStruct.Field(i).Set(getTupleToStruct(Data,TheOtherX,sEMId,tableName+"_"+s[1]))



		}else if v=="Id"{//primary key or foreign key

			//reflect.ValueOf(newStruct).Field(i).Set(reflect.ValueOf(int(slice[i].(int64))))


		}else{//normal

			s:=strings.Split(v,"_")
			switch s[1] {
			case "int":
				val:=reflect.ValueOf(int(cachePS[i+1].(int64)))
				_=val
				newStruct.Field(i).Set(val)
			case "string":
				val:=reflect.ValueOf(cachePS[i+1].(string))
				newStruct.Field(i).Set(val)

			case "float":
				val:=reflect.ValueOf(float32(cachePS[i+1].(float64)))
				newStruct.Field(i).Set(val)

			default:


			}


		}

	}



	return newStruct


}






func getTupleCount(Data *Database,tableName string) int{
	var result int

	q1:=fmt.Sprintf("Select Count(rowid) FROM %s",tableName)
	row:=Data.db.QueryRow(q1)
	row.Scan(&result)

	return result

}

func getTupleCountWithId(Data *Database,tableName string,Id int)int{
	var result int
	q1:="Select Count(*) FROM "+tableName+" WHERE Id=="+strconv.Itoa(Id)
	row:=Data.db.QueryRow(q1)
	row.Scan(&result)
	return result
}

func getColumnNames(Data *Database,tableName string) []string{

	var all []string
	var al string
	var count int
	q1:=fmt.Sprintf("Select Count(name) FROM PRAGMA_table_info('%s');",tableName)
	row:=Data.db.QueryRow(q1)
	row.Scan(&count)

	q2:=fmt.Sprintf("Select name FROM PRAGMA_table_info('%s');",tableName)
	rows,err:=Data.db.Query(q2)
	Check(err)
	for rows.Next(){
		err:=rows.Scan(&al)
		Check(err)
		all=append(all,al)
	}
	defer rows.Close()

	return  all

}
func newStructLike(Astruct interface{}) (newStruct interface{}){ //this method return settable and adressable struct
	ptr:=reflect.New(reflect.TypeOf(Astruct))
	newStruct=ptr.Elem()
	return newStruct
}

func getAllTableNamesforDb(Data *Database){
	var names []string
	var name string
	q:="SELECT name FROM sqlite_master WHERE type ='table' AND name NOT LIKE 'sqlite_%';"
	rows,err:=Data.db.Query(q)
	Check(err)
	for rows.Next(){
		rows.Scan(&name)
		names=append(names,name)
	}
	defer rows.Close()

	Check(err)
	Data.AllTables=names
}

func execTheQueries(data *Database,Q []string){

	for _,v:=range Q{
		res,err:=data.db.Exec(v)
		_=res
		Check(err)
	}

}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}


func getQueries(t Table) []string{

	var queries []string

	var emptytt []Table
	var tt = t
	for tt.ContainsArr||tt.ContainsEmbedded{
		emptytt=emptytt[:0]
		emptytt=append(emptytt,tt.ArrTables...)
		emptytt=append(emptytt,tt.EmbeddedTables...)
		for _,v:=range emptytt{
			queries=append(queries,getQuery(v))
			tt=v
		}

	}

	var ttt=t
	var emptytt2 []Table
	for ttt.ContainsEmbedded||ttt.ContainsArr{
		emptytt2=emptytt2[:0]
		emptytt2=append(emptytt2,ttt.EmbeddedTables...)
		emptytt2=append(emptytt2,ttt.ArrTables...)
		for _,v:=range emptytt2{
			queries=append(queries,getQuery(v))
			ttt=v
		}
	}

	queries=append(queries,getQuery(t))

	queries=unique(queries)

	return queries

}

func getQuery(t Table) string{
	var Query string = "CREATE TABLE IF NOT EXISTS "+t.Name+"("

	for i:=0;i< len(t.Columns);i++{//range over the columns name

		if i==0{
			Query+=t.Columns[i]+" "+getType(t.ColTypes[i])
		}else{
			Query+=","+t.Columns[i]+" "+getType(t.ColTypes[i])
		}




	}
	Query+=")"

	return Query

}

func getType(typs reflect.Type) string{

	switch typs.Kind() {

	case reflect.Int,reflect.Int8,reflect.Uint16,reflect.Int32,reflect.Int64:
		return "INTEGER"
	case reflect.Float32,reflect.Float64:
		return "REAL"
	case reflect.String:
		return "TEXT"
	case reflect.Struct:
		return "INTEGER" //its for foreign key
	default:
		return "TEXT"
	}

}

func getTable2(colname,tableName string,typ reflect.Type) Table{

	newTable:=Table{
		Name:tableName,
		Columns:[]string{"Id",colname},
		ContainsArr:false,
		ArrTables:nil,
		ContainsEmbedded:false,
		ColTypes:[]reflect.Type{reflect.TypeOf(0),typ},
		EmbeddedTables:nil,


	}

	return newTable
}

func getTable(x interface{},tableName string) Table{


	value:=reflect.ValueOf(x)

	colnames:=fieldNames(x)
	coltyps:=fieldTypes(x)

	newTable:=Table{
		Name:tableName,
		ContainsEmbedded:false,
		ContainsArr:false,

	}

	var newColnames []string
	var newColtyps []reflect.Type

	newColnames=append(newColnames,"Id")
	newColtyps=append(newColtyps,reflect.TypeOf(0))//for integer type!

	for i,v:=range coltyps{
		if v.Kind()==reflect.Struct{
			newColnames=append(newColnames,"EM_"+colnames[i]) //EM_ for embedded
			newColtyps=append(newColtyps,reflect.TypeOf(0))
			newTable.EmbeddedTables=append(newTable.EmbeddedTables,getTable(value.Field(i).Interface(),tableName+"_"+value.Type().Field(i).Name))
			newTable.ContainsEmbedded=true
			//added to table (parent)




		}else if v.Kind()==reflect.Array||v.Kind()==reflect.Slice{
			//newColnames=append(newColnames,"EMARR_"+colnames[i])//EMARR_ for slice or array types
			newColtyps=append(newColtyps,reflect.TypeOf(0))//we store it as key to create relation its table
			newTable.ContainsArr=true
			if value.Field(i).Type().Elem().Kind()==reflect.Struct{//if the slice(or array) contain struct
				newColnames=append(newColnames,"EMARR_AN_"+colnames[i])//EMARR_ for slice or array types
				newTable.ArrTables=append(newTable.ArrTables,getTable(reflect.Zero(value.Field(i).Type().Elem()).Interface(),tableName+"_"+value.Type().Field(i).Name))


			}else{//it is not struct
				//we need a new function in this area , we do not need recursive anymore
				newColnames=append(newColnames,"EMARR_"+value.Field(i).Type().Elem().Name()+"_"+colnames[i])//EMARR_ for slice or array types
				newTable.ArrTables=append(newTable.ArrTables,getTable2(value.Type().Field(i).Name+"_"+value.Field(i).Type().Elem().Name(),tableName+"_"+value.Type().Field(i).Name,value.Field(i).Type().Elem()))

			}

		}else{

			//for normal types
			newColnames=append(newColnames,colnames[i]+"_"+value.Field(i).Type().Name())
			newColtyps=append(newColtyps,v)

		}
	}

	newTable.Columns=newColnames
	newTable.ColTypes=newColtyps

	return newTable

}

func fieldNames(Astruct interface{}) (fieldNames []string){
	value:=reflect.ValueOf(Astruct)
	for i:=0;i<value.NumField();i++{
		fieldNames=append(fieldNames,value.Type().Field(i).Name)
	}

	return
}

func fieldTypes(Astruct interface{}) (types []reflect.Type){
	value:=reflect.ValueOf(Astruct)

	for i:=0;i<value.NumField();i++{
		types=append(types,value.Type().Field(i).Type)
	}

	return types
}

func NewSliceFor(item interface{},len,cap int)(newSlice interface{}){
	newSlice=reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(item)),len,cap)
	return newSlice
}

func (c *Collection) Add(x interface{}) {

	if c.lock==false {

		if len(c.LocalH) > 0 {

			//local hook
			var beforeAddLocalHooks []LocalHook
			for i := 0; i < len(c.LocalH); i++ {
				if _, n := c.LocalH[i].getSign(); n == BeforeAdd {
					beforeAddLocalHooks = append(beforeAddLocalHooks, c.LocalH[i])
				}
			}

			sort.Sort(byPriority(beforeAddLocalHooks))

			for i := 0; i < len(beforeAddLocalHooks); i++ {
				funk := beforeAddLocalHooks[i].getHookFunc()
				x = funk(x)
			}
			//local hook

		}
	}


	keys:=getKeysAsSlice(c.List)
	var newID int
	if len(keys)==0{
		newID=1
	}else{
		newID=keys[len(keys)-1]+1
	}
	//db insert start

	setTuple(x,c.ITsTable.Name, newID,*c)

	//db insert

	if len(keys)==0{
		c.List[1]=x
	}else{
		c.List[newID]=x //inserted!
	}

	if c.lock==false {
		if len(c.LocalH) > 0 {
			//local hook
			var afterAddLocalHooks []LocalHook
			for i := 0; i < len(c.LocalH); i++ {
				if _, n := c.LocalH[i].getSign(); n == AfterAdd {
					afterAddLocalHooks = append(afterAddLocalHooks, c.LocalH[i])
				}
			}

			sort.Sort(byPriority(afterAddLocalHooks))

			for i := 0; i < len(afterAddLocalHooks); i++ {
				funk := afterAddLocalHooks[i].getHookFunc()
				funk(x)
			}
			//local hook

		}
	}

}


func (c *Collection) Delete(obj interface{}) {

	if c.lock==false {
		if len(c.LocalH) > 0 {
			//local hook
			var beforeDeleteLocalHooks []LocalHook
			for i := 0; i < len(c.LocalH); i++ {
				if _, n := c.LocalH[i].getSign(); n == BeforeDelete {
					beforeDeleteLocalHooks = append(beforeDeleteLocalHooks, c.LocalH[i])
				}
			}

			sort.Sort(byPriority(beforeDeleteLocalHooks))

			for i := 0; i < len(beforeDeleteLocalHooks); i++ {
				funk := beforeDeleteLocalHooks[i].getHookFunc()
				obj = funk(obj)
			}
			//local hook

		}
	}

	for k,v:=range c.List{
		if reflect.DeepEqual(v,obj){
			//db delete start
			deleteTuple(c.ITsTable,k,*c)
			//delete finish
			delete(c.List,k)//delete item from map
			break
		}
	}


	if c.lock==false {
		if len(c.LocalH) > 0 {

			//local hook
			var afterDeleteLocalHooks []LocalHook
			for i := 0; i < len(c.LocalH); i++ {
				if _, n := c.LocalH[i].getSign(); n == AfterDelete {
					afterDeleteLocalHooks = append(afterDeleteLocalHooks, c.LocalH[i])
				}
			}

			sort.Sort(byPriority(afterDeleteLocalHooks))

			for i := 0; i < len(afterDeleteLocalHooks); i++ {
				funk := afterDeleteLocalHooks[i].getHookFunc()
				funk(obj)
			}
			//local hook

		}
	}



}

func (c *Collection) Clear(){

	deleteTable(c.ITsTable,*c)
	c.List=make(map[int]interface{})

}

func (c *Collection) Update(old interface{},new interface{}){

    c.lock=true
	if len(c.LocalH)>0 {
		//local hook
		var beforeUpdateLocalHooks []LocalHook
		for i := 0; i < len(c.LocalH); i++ {
			if _, n := c.LocalH[i].getSign(); n == BeforeUpdate {
				beforeUpdateLocalHooks = append(beforeUpdateLocalHooks, c.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeUpdateLocalHooks))

		for i := 0; i < len(beforeUpdateLocalHooks); i++ {
			funk := beforeUpdateLocalHooks[i].getHookFunc()
			old = funk(old)
		}
		//local hook

	}

	var Ids []int

	for k,v:=range c.List{

		if reflect.DeepEqual(v,old){
			Ids=append(Ids,k) //this Ids will be update in database side
			c.List[k]=new //update completed!
		}

	}

	//Update start
	if len(Ids)>0{ //check

		updateTuple(new,c.ITsTable.Name,Ids,*c)

	}
	//update finish

	if len(c.LocalH)>0 {
		//local hook
		var afterUpdateLocalHooks []LocalHook
		for i := 0; i < len(c.LocalH); i++ {
			if _, n := c.LocalH[i].getSign(); n == AfterUpdate {
				afterUpdateLocalHooks = append(afterUpdateLocalHooks, c.LocalH[i])
			}
		}

		sort.Sort(byPriority(afterUpdateLocalHooks))

		for i := 0; i < len(afterUpdateLocalHooks); i++ {
			funk := afterUpdateLocalHooks[i].getHookFunc()
			funk(new)
		}
		//local hook

	}
	c.lock=false

}

func updateTuple(new interface{},tableName string,Ids []int,collection Collection){//it uses insert and delete we'll replace with recursive update in next version

	for _,Id:=range Ids{
		deleteTuple(collection.ITsTable,Id,collection)
	}
	for _,Id:=range Ids{
		setTuple(new,tableName,Id,collection)
	}


}

func (c Collection) ToSlice() []interface{}{
	slice:=make([]interface{},0,1)

	for _,v:=range c.List{
		slice=append(slice,v)
	}

	return slice
}

/*func getUpdateQuery(olds map[int]interface{},diff map[int]interface{},tableName string, Id int, collection Collection) string{
	colNames:=getColumnNames(collection.data,tableName)
	var Query string
	Query="UPDATE "+tableName+" SET "

	for k,_:=range diff{
		Query+=""+colNames[k+1]+"=? ,"
	}

	Query+=" WHERE Id=? AND "

	keys:=getKeysAsSlice(olds)

	for k,_:=range olds{
		if keys[len(keys)-1]==k{
			Query+=" "+colNames[k+1]+"=? "
		}else{
			Query+=" "+colNames[k+1]+"=? AND"
		}

	}

	return Query

}*/




/*func getUpdateQuery(t Table,Id int,diff map[int]interface{},collection Collection) []string{
	tables:=getAllTableNames(t)
	var queries []string
	var query string

	for _,v :=range tables{
		columns:=getColumnNames(collection.data,v)
		columns=columns[1:]
		query="UPDATE "+v+" SET "
		for i,_:=range diff{
			if i==0{
				query+=columns[i]+"=?"
			}else{
				query+=","+columns[i]+"=?"
			}

		}

		query+=" WHERE Id==?"


        queries=append(queries,query)

	}

	return queries

}*/

func deleteTable(t Table,collection Collection){
	tables:=getAllTableNames(t)

	for _,v:=range tables{

		res,err:=collection.data.db.Exec("DELETE FROM  "+v)


		Check(err)

		_=res
	}
}

func deleteTuple(t Table,Id int,collection Collection){


	tables:=getAllTableNames(t)

	for _,v:=range tables{


		res,err:=collection.data.db.Exec("DELETE FROM  "+v+" WHERE Id==?",Id)


		Check(err)

		_=res
	}


}

func getAllTableNames(t Table) []string{
	var queries []string

	var emptytt []Table
	var tt = t
	for tt.ContainsArr||tt.ContainsEmbedded{
		emptytt=emptytt[:0]
		emptytt=append(emptytt,tt.ArrTables...)
		emptytt=append(emptytt,tt.EmbeddedTables...)
		for _,v:=range emptytt{
			queries=append(queries,v.Name)
			tt=v
		}

	}

	var ttt=t
	var emptytt2 []Table
	for ttt.ContainsEmbedded||ttt.ContainsArr{
		emptytt2=emptytt2[:0]
		emptytt2=append(emptytt2,ttt.EmbeddedTables...)
		emptytt2=append(emptytt2,ttt.ArrTables...)
		for _,v:=range emptytt2{
			queries=append(queries,v.Name)
			ttt=v
		}
	}

	queries=append(queries,t.Name)

	queries=unique(queries)

	return queries
}

func getMaxColumnValue(Data *Database,columnName,tableName string) int{
	var result int
	row:=Data.db.QueryRow("SELECT MAX("+columnName+") FROM "+tableName)
	err:=row.Scan(&result)
	if err!=nil{
		return 0
	}
	Check(err)
	return result
}

func setTuples(slice interface{},tableName string,Id int,collection Collection){


	sl:=reflect.ValueOf(slice)

	fmt.Println(slice)
	fmt.Println(sl)
	fmt.Println(getInsertQuery(tableName, getColumnNames(collection.data, tableName)))
	Tx,err:=collection.data.db.Begin()

	Check(err)
	var theeId = 0
	Idchecked:=false
	for i:=0;i<sl.Len();i++ {

		stmt, err := Tx.Prepare(getInsertQuery(tableName, getColumnNames(collection.data, tableName)))

		Check(err)
        bak:=getColumnNames(collection.data, tableName)
        slc:=make([]interface{},0)
        slc=append(slc,Id)
        for k:=0;k< len(bak)-1;k++{



        	if sl.Index(i).Kind()==reflect.Array||sl.Index(i).Kind()==reflect.Slice{
				if !Idchecked{
					theId:=getMaxColumnValue(collection.data,bak[k+1],tableName)+1
					theeId=theId
					Idchecked=true
				}
				s:=strings.Split(bak[k+1],"_")

				setTuple2(sl.Index(i).Interface(),tableName+"_"+s[len(s)-1],theeId,collection,[]string{"Id",tableName+"_"+s[len(s)-1]})
				return
			}

        	switch sl.Index(i).Field(k).Kind(){

			case reflect.Slice,reflect.Array,reflect.Struct://burada int,float,string vs ayrı ayrı handle etmek gerekebilir
				if !Idchecked{
					theId:=getMaxColumnValue(collection.data,bak[k+1],tableName)+1
					theeId=theId
					Idchecked=true
				}
				slc=append(slc,theeId)
				theeId++;
				s:=strings.Split(bak[k+1],"_")
				setTuples(sl.Index(i).Field(k).Interface(),tableName+"_"+s[len(s)-1],theeId-1,collection)

				/*switch sl.Index(i).Field(k).Elem().Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
					setTuple(int(0),tableName+"_"+s[len(s)-1],theeId-1,collection)
				case reflect.Float32,reflect.Float64:
					setTuple(float32(0),tableName+"_"+s[len(s)-1],theeId-1,collection)
				case reflect.String:
					setTuple(string("0"),tableName+"_"+s[len(s)-1],theeId-1,collection)
				case reflect.Bool:

				default:


				}*/

			default:
				slc=append(slc,sl.Index(i).Field(k).Interface())//sql-engine issue here!!!
			}
			fmt.Println(slc)


		}
		res, err := stmt.Exec(slc...)
		Check(err)
		_ = res

		if err != nil {
			err = Tx.Rollback()
			Check(err)
		}




		err=stmt.Close()
		Check(err)
	}
	err = Tx.Commit()
	Check(err)





}
func setTuple(x interface{},tableName string,Id int,collection Collection){
	value:=reflect.ValueOf(x)
	colNames:=getColumnNames(collection.data,tableName)
	myMap:=make(map[string]interface{})

	for i,v:=range colNames{
		if strings.Contains(v,"EMARR_"){//array or slice field!
			s:=strings.Split(v,"_")
			setTuples(value.Field(i-1).Interface(),tableName+"_"+s[2],Id,collection)
			myMap[v]=Id //like foreign key
			/*s:=strings.Split(v,"_")
			myMap[v]=Id //like foreign key
			setTuple(value.Field(i-1).Interface(),tableName+"_"+s[2],Id,collection)*/

		}else if strings.Contains(v,"EM_"){//for embedded field
			s:=strings.Split(v,"_")
			myMap[v]=Id//like foreign key
			setTuple(value.Field(i-1).Interface(),tableName+"_"+s[1],Id,collection)

		}else if v=="Id"{

			myMap[v]=Id

		}else{//for normal types

			myMap[v]=value.Field(i-1).Interface()
		}
	}

	Tx,err:=collection.data.db.Begin()
	Check(err)
	stmt,err:=Tx.Prepare(getInsertQuery(tableName,colNames))
	Check(err)
	giveMeaSlice:=orderForInsert(myMap,colNames)
	res,err:=stmt.Exec(giveMeaSlice...)//passed
	if err!=nil{
		err=Tx.Rollback()
		Check(err)
	}

	err=Tx.Commit()

	Check(err)
	err=stmt.Close()
	Check(err)

	_=res
	_=Tx
}

func setTuple2(x interface{},tableName string,Id int,collection Collection,colNames []string){
	value:=reflect.ValueOf(x)
	//colNames:=getColumnNames(collection.data,tableName)
	myMap:=make(map[string]interface{})

	for i,v:=range colNames{
		if strings.Contains(v,"EMARR_"){//array or slice field!
			s:=strings.Split(v,"_")
			setTuples(value.Field(i-1).Interface(),tableName+"_"+s[2],Id,collection)
			myMap[v]=Id //like foreign key
			/*s:=strings.Split(v,"_")
			myMap[v]=Id //like foreign key
			setTuple(value.Field(i-1).Interface(),tableName+"_"+s[2],Id,collection)*/

		}else if strings.Contains(v,"EM_"){//for embedded field
			s:=strings.Split(v,"_")
			myMap[v]=Id//like foreign key
			setTuple(value.Field(i-1).Interface(),tableName+"_"+s[1],Id,collection)

		}else if v=="Id"{

			myMap[v]=Id

		}else{//for normal types

			myMap[v]=value.Field(i-1).Interface()
		}
	}

	Tx,err:=collection.data.db.Begin()
	Check(err)
	stmt,err:=Tx.Prepare(getInsertQuery(tableName,colNames))
	Check(err)
	giveMeaSlice:=orderForInsert(myMap,colNames)
	res,err:=stmt.Exec(giveMeaSlice...)//passed
	if err!=nil{
		err=Tx.Rollback()
		Check(err)
	}

	err=Tx.Commit()

	Check(err)
	err=stmt.Close()
	Check(err)

	_=res
	_=Tx
}


func orderForInsert(myMap map[string]interface{},cols []string) []interface{}{
	slice:=make([]interface{},0,1)
	for _,v:=range cols{
		slice=append(slice,myMap[v])
	}
	return slice
}


func getInsertQuery(tableName string , Columns []string) string{

	var Query string = "INSERT INTO "+tableName+"("

	for i:=0;i< len(Columns);i++{//range over the columns name

		if i==0{
			Query+=Columns[i]+" "
		}else{
			Query+=","+Columns[i]+" "
		}




	}
	Query+=")"

	Query+=" VALUES ("
	for i:=0;i< len(Columns);i++{
		if i==0{
			Query+="? "
		}else{
			Query+=","+"? "
		}
	}
	Query+=" )"

	return Query
}



/*func insertTuple(x interface{},tableName string,Id int){
	value:=reflect.ValueOf(x)

	colnames:=fieldNames(x)
	coltyps:=fieldTypes(x)

	cache:=make([]interface{}, len(colnames)+1, len(colnames)+1)

	cache=append(cache,Id)//firstly add Id

	for i,v:=range coltyps{
		if v.Kind()==reflect.Struct{



		}else if v.Kind()==reflect.Array||v.Kind()==reflect.Slice{
		//EMARR_ for slice or array types

			if value.Field(i).Type().Elem().Kind()==reflect.Struct{//if the slice(or array) contain struct

			}else{//it is not struct


			}

		}else{//for normal types

         cache[i+1]=reflect.ValueOf(x).Field(i).Interface()


		}
	}




}

/*func insertStruct(x interface{},collection Collection){
	TheStruct:=reflect.ValueOf(x)

	var queries []string
	tables:=getInsertTables(collection.ITsTable)
	for _,v:=range tables{
		queries=append(queries,getInsertQuery(v))
	}

}

func getInsertTables(t Table) []Table {
	var queries []Table

	var emptytt []Table
	var tt = t
	for tt.ContainsArr||tt.ContainsEmbedded{
		emptytt=emptytt[:0]
		emptytt=append(emptytt,tt.ArrTables...)
		emptytt=append(emptytt,tt.EmbeddedTables...)
		for _,v:=range emptytt{
			queries=append(queries,v)
			tt=v
		}

	}

	var ttt=t
	var emptytt2 []Table
	for ttt.ContainsEmbedded||ttt.ContainsArr{
		emptytt2=emptytt2[:0]
		emptytt2=append(emptytt2,ttt.EmbeddedTables...)
		emptytt2=append(emptytt2,ttt.ArrTables...)
		for _,v:=range emptytt2{
			queries=append(queries,v)
			ttt=v
		}
	}

	queries=append(queries,t)

	queries=unique2(queries)

	return queries
}

func getInsertQuery(t Table) string{

	var Query string = "INSERT INTO "+t.Name+"("

	for i:=0;i< len(t.Columns);i++{//range over the columns name

		if i==0{
			Query+=t.Columns[i]+" "
		}else{
			Query+=","+t.Columns[i]+" "
		}




	}
	Query+=")"

	Query+=" VALUES ("
	for i:=0;i< len(t.Columns);i++{
		if i==0{
			Query+="? "
		}else{
			Query+=","+"? "
		}
	}
	Query+=" )"

	return Query
}*/

func getKeysAsSlice(mp map[int]interface{}) []int{
	keys:=make([]int,0,1)

	for k:=range mp {
		keys=append(keys,k)
	}

	sort.Ints(keys)

	return keys
}




func(c *Collection) Foreach(interface{}){//!!!must be supported

}
func(c *Collection) AddRange(interface{}){//!!!must be supported

}
func(c *Collection) GetLogs(){//!!!must be supported

}



