package Orca

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type MysqlDB struct {
	Name string
	Database *sql.DB
	ConnectionString string
	LocalH []LocalHook
}


type MysqlCollection struct {
	Mysql *MysqlDB
	ListId []int
	List []interface{}
	LocalH []LocalHook
	ConnectionString string
	DatabaseName string
	Config MySqlOrmConfig
	CacheList map[int]interface{}
	lock bool
}


func getMysqlDB(user, password, dbName, network string) *MysqlDB{

	db,err:=sql.Open("mysql",user+":"+password+"@"+network+"/")
	Check(err)

	_,err=db.Exec("CREATE DATABASE IF NOT EXISTS "+dbName)


	_,err=db.Exec("use "+dbName)
	Check(err)


	//db,err=sql.Open("mysql",connString)
	//Check(err)

	return &MysqlDB{
		Name:dbName,
		Database:db,
		ConnectionString:user+":"+password+"@"+network+"/"+dbName,
		LocalH: []LocalHook{},
	}
}

func fixTheQueriesMySQl(queries []string,mainTable string) []string{
	for i:=0;i< len(queries);i++{
		//if !strings.Contains(queries[i],"TABLE_NAME = N'"+mainTable+"')"){
		if !strings.Contains(queries[i],"ForeignID"){
			queries[i]=strings.Replace(queries[i],"(Id int(11) NOT NULL auto_increment,","(Id int(11) NOT NULL auto_increment,ForeignID int(11) NULL,",1)
		}
		//}
	}
	return  queries
}

func executeQueriesMySql(mysql *MysqlDB, queries []string){
	for i:=0;i< len(queries);i++{
		_,err:=mysql.Database.Exec(queries[i])
		Check(err)
	}
}

func (mysql *MysqlDB) GetCollection(x interface{}, tableName string) ICollection {

	_,err:=mysql.Database.Exec("CREATE TABLE IF NOT EXISTS OrcaConfig(Id int(11) NOT NULL auto_increment,TableName varchar(255) NULL,ColumnName varchar(255) NULL,ColumnType varchar(255) NULL,ColumnType2 varchar(50) NULL,ColumnOrder int(11) NULL, PRIMARY KEY(Id));")
	Check(err)

	_,err=mysql.Database.Exec("CREATE TABLE IF NOT EXISTS OrcaRelations(Id int(11)  NOT NULL auto_increment,Parent varchar(255) NULL,Child varchar(255) NULL,Slice bit NOT NULL,Array bit NOT NULL,Embedded BIT NOT NULL,PRIMARY KEY(Id));")
	Check(err)

	var count int
	row:=mysql.Database.QueryRow("SELECT COUNT(*) FROM OrcaConfig Where TableName=?",tableName)
	err=row.Scan(&count)
	Check(err)

	getTheDBSchema:=createOrmConfigForMySQl(x,tableName)

	allSQLQueries := fixTheQueriesMySQl(getTheDBSchema.toSQL(mysql),tableName)

	var list []interface{}
	var listID []int
	var cacheList map[int]interface{} = make(map[int]interface{})

	if count==0{
		executeQueriesMySql(mysql,allSQLQueries)
		createConfigAndRelationsReferancesForDatabaseMySql(mysql,getTheDBSchema)

	}else{
		theMin:=getMinForeignIDMySQl(mysql,getTheDBSchema.TableName)
		TheMax:=getMaxForeignIDMySQl(mysql,getTheDBSchema.TableName)

		if TheMax-theMin>=0{
			for i:=theMin;i<=TheMax;i++{
				a:=getTheDBSchema.readStructFromDB(i,mysql,x,false,0)
				list=append(list,a)
				listID=append(listID,i)
				cacheList[i]=a
			}
		}

	}

    return &MysqlCollection{
		Mysql:            mysql,
		ListId:           listID,
		List:             list,
		LocalH:           mysql.LocalH,
		ConnectionString: mysql.ConnectionString,
		DatabaseName:     mysql.Name,
		Config:           getTheDBSchema,
		CacheList:        cacheList,
		lock:             false,
	}
}

func getMinForeignIDMySQl(mysql *MysqlDB, tablename string) int{
	var result int
	row:=mysql.Database.QueryRow("Select Min(ForeignID) FROM "+tablename)
	err:=row.Scan(&result)
	Check(err)
	if err!=nil{
		result=0
	}
	return result
}

func getMaxForeignIDMySQl(mysql *MysqlDB, tablename string) int{
	var result int
	row:=mysql.Database.QueryRow("Select Max(ForeignID) FROM "+tablename)
	err:=row.Scan(&result)
	Check(err)
	if err!=nil{
		result=0
	}
	return result
}

type MySqlOrmConfig struct {
	TableName string
	TypeAnalysis map[int]string
	CinsAnalysis map[int]string
	ColumnOrder map[string]int
	ColumnNames map[int]string
	AltTable map[int]string
	EMObjects map[string]reflect.Value
	AltTableObj []MySqlOrmConfig
	ContainsEM bool
	ContainsEMTablesName []string
	ContainsEMARR bool
	ContainsEMARRTablesName []string
	ContainsEMSLICE bool
	ContainsEMSLICETablesName []string
	ContainsUNDEFINED bool
}

func(config MySqlOrmConfig) getAltTableByName(name string) MySqlOrmConfig{
	dz:=config.getAllAltTables()
	for i:=0;i< len(dz);i++{
		if dz[i].TableName==name{
			return dz[i]
		}
	}
	return MySqlOrmConfig{}
}

func getCountOfTuplesMySQl(id int,mysql *MysqlDB,tableName string) int{
	var result int = 0
	row:=mysql.Database.QueryRow("Select Count(*) from " +tableName + " WHERE ForeignID=?",id)
	row.Scan(&result)
	return  result
}

func(config MySqlOrmConfig) readStructFromDB(Id int,mysql *MysqlDB,x interface{},isRecursiveCall bool,recNumber int) interface{}{
	var allOfThem []interface{}
	if !isRecursiveCall{
		recNumber=0
	}
	ptr:=reflect.New(reflect.TypeOf(x))
	value:=ptr.Elem()
	for i:=0;i<value.NumField();i++{
		typeOfValue:=config.TypeAnalysis[i]
		switch typeOfValue {
		case "NORMAL":
			switch config.CinsAnalysis[i] {
			case "string":
				if isRecursiveCall{
					var strArr []string
					var str string
					rows,err:=mysql.Database.Query("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					Check(err)
					defer rows.Close()
					for rows.Next(){
						rows.Scan(&str)
						strArr=append(strArr,str)
					}
					allOfThem=append(allOfThem,strArr)
				}else{
					var str string
					row:=mysql.Database.QueryRow("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					row.Scan(&str)
					value.Field(i).SetString(str)
				}
			case "int":
				if isRecursiveCall{
					var intARR []int64
					var integer int
					rows,err:=mysql.Database.Query("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					Check(err)
					defer rows.Close()
					for rows.Next(){
						rows.Scan(&integer)
						intARR=append(intARR,int64(integer))
					}
					allOfThem=append(allOfThem,intARR)
				}else{
					var integer int
					row:=mysql.Database.QueryRow("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					row.Scan(&integer)
					value.Field(i).SetInt(int64(integer))
				}
			case "bool":
				if isRecursiveCall{
					var boolARR []bool
					var tf bool
					rows,err:=mysql.Database.Query("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					Check(err)
					defer rows.Close()
					for rows.Next(){
						rows.Scan(&tf)
						boolARR=append(boolARR,tf)
					}
					allOfThem=append(allOfThem,boolARR)
				}else{
					var tf bool
					row:=mysql.Database.QueryRow("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					row.Scan(&tf)
					value.Field(i).SetBool(tf)
				}
			case "float32","float":
				if isRecursiveCall{
					var floatARR []float64
					var fl float32
					rows,err:=mysql.Database.Query("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					Check(err)
					defer rows.Close()
					for rows.Next(){
						rows.Scan(&fl)
						floatARR=append(floatARR,float64(fl))
					}
					allOfThem=append(allOfThem,floatARR)
				}else{
					var fl float32
					row:=mysql.Database.QueryRow("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					row.Scan(&fl)
					value.Field(i).SetFloat(float64(fl))
				}
			case "float64":
				if isRecursiveCall{
					var floatARR []float64
					var fl float64
					rows,err:=mysql.Database.Query("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					Check(err)
					defer rows.Close()
					for rows.Next(){
						rows.Scan(&fl)
						floatARR=append(floatARR,fl)
					}
					allOfThem=append(allOfThem,floatARR)
				}else{
					var fl float64
					row:=mysql.Database.QueryRow("SELECT "+config.ColumnNames[i]+" FROM "+config.TableName+" WHERE ForeignID=?",Id)
					row.Scan(&fl)
					value.Field(i).SetFloat(fl)
				}
			default:

			}
		case "SLICE":
			switch strings.Split(config.CinsAnalysis[i],"_")[1] {
			case "string":
				var stringARRorSLI []string
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					str:=""
					rows.Scan(&str)
					stringARRorSLI=append(stringARRorSLI,str)
				}
				value.Field(i).Set(reflect.ValueOf(stringARRorSLI))
				allOfThem=append(allOfThem,stringARRorSLI)
			case "int":
				var intARRorSLI []int
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var integer int
					rows.Scan(&integer)
					intARRorSLI=append(intARRorSLI,integer)
				}
				value.Field(i).Set(reflect.ValueOf(intARRorSLI))
				allOfThem=append(allOfThem,intARRorSLI)
			case "bool":
				var boolARRorSLI []bool
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var bl bool
					rows.Scan(&bl)
					boolARRorSLI=append(boolARRorSLI,bl)
				}
				value.Field(i).Set(reflect.ValueOf(boolARRorSLI))
				allOfThem=append(allOfThem,boolARRorSLI)
			case "float32","float":
				var floatARRorSLI []float32
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var fl float32
					rows.Scan(&fl)
					floatARRorSLI=append(floatARRorSLI,fl)
				}
				value.Field(i).Set(reflect.ValueOf(floatARRorSLI))
				allOfThem=append(allOfThem,floatARRorSLI)
			case "float64":
				var floatARRorSLI []float64
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var fl float64
					rows.Scan(&fl)
					floatARRorSLI=append(floatARRorSLI,fl)
				}
				value.Field(i).Set(reflect.ValueOf(floatARRorSLI))
				allOfThem=append(allOfThem,floatARRorSLI)
			default:

			}
		case "ARR":
			switch strings.Split(config.CinsAnalysis[i],"_")[1] {
			case "string":
				var stringARRorSLI []string
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					str:=""
					rows.Scan(&str)
					stringARRorSLI=append(stringARRorSLI,str)
				}
				for ar:=0;ar<value.Field(i).Len();ar++{
					value.Field(i).Index(ar).SetString(stringARRorSLI[ar])
				}
				allOfThem=append(allOfThem,stringARRorSLI)
			case "int":
				var intARRorSLI []int
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var integer int
					rows.Scan(&integer)
					intARRorSLI=append(intARRorSLI,integer)
				}
				for ar:=0;ar<value.Field(i).Len();ar++{
					value.Field(i).Index(ar).SetInt(int64(intARRorSLI[ar]))
				}
				var intARR []int64
				for _,v:= range intARRorSLI{
					intARR=append(intARR,int64(v))
				}
				allOfThem=append(allOfThem,intARR)
			case "bool":
				var boolARRorSLI []bool
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var bl bool
					rows.Scan(&bl)
					boolARRorSLI=append(boolARRorSLI,bl)
				}
				for ar:=0;ar<value.Field(i).Len();ar++{
					value.Field(i).Index(ar).SetBool(boolARRorSLI[ar])
				}
				allOfThem=append(allOfThem,boolARRorSLI)
			case "float32","float":
				var floatARRorSLI []float32
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var fl float32
					rows.Scan(&fl)
					floatARRorSLI=append(floatARRorSLI,fl)
				}
				for ar:=0;ar<value.Field(i).Len();ar++{
					value.Field(i).Index(ar).SetFloat(float64(floatARRorSLI[ar]))
				}
				var floatARR []float64
				for _,v:=range floatARRorSLI{
					floatARR=append(floatARR,float64(v))
				}
				allOfThem=append(allOfThem,floatARR)
			case "float64":
				var floatARRorSLI []float64
				rows,err:=mysql.Database.Query("SELECT Value FROM "+config.TableName+"_"+strconv.Itoa(i)+" WHERE ForeignID=?",Id)
				Check(err)
				defer rows.Close()
				for rows.Next(){
					var fl float64
					rows.Scan(&fl)
					floatARRorSLI=append(floatARRorSLI,fl)
				}
				for ar:=0;ar<value.Field(i).Len();ar++{
					value.Field(i).Index(ar).SetFloat(float64(floatARRorSLI[ar]))
				}
				var floatARR []float64
				for _,v:=range floatARRorSLI{
					floatARR=append(floatARR,float64(v))
				}
				allOfThem=append(allOfThem,floatARR)
			default:

			}
		case "EMARR":

			underlyingValue:=value.Field(i).Type().Elem()
			olusturulanConfigBu := createOrmConfigForMySQl(reflect.New(underlyingValue).Elem().Interface(),config.TableName+"_"+strconv.Itoa(i))
			gelen:=reflect.ValueOf(olusturulanConfigBu.readStructFromDB(Id,mysql,reflect.New(underlyingValue).Elem().Interface(),true, value.Field(i).Len())) //it returns a slice, each element for a column

			for turn:=0;turn<value.Field(i).Len();turn++{

				for lol:=0;lol<gelen.Len();lol++{
					switch olusturulanConfigBu.TypeAnalysis[lol] {
					case "NORMAL":
						value.Field(i).Index(turn).Field(lol).Set(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol).Index(turn))
					case "ARR":
						for lollol:=0;lollol<gelen.Index(lol).Len();lollol++{
							value.Field(i).Index(turn).Field(lol).Index(lollol).Set(gelen.Index(lol).Index(lollol))
						}
						/*for ii,vv:=range v{
							value.Field(i).Index(turn).Field(lol).Index(ii).Set(vv)
						}*/
					case "SLICE":
						value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol))
					case "EMARR":
						for lollol:=0;lollol<gelen.Index(lol).Len();lollol++{
							value.Field(i).Index(turn).Field(lol).Index(lollol).Set(gelen.Index(lol).Index(lollol))
						}
						/*for ii,vv:=range v{
							value.Field(i).Index(turn).Field(lol).Index(ii).Set(vv)
						}*/
					case "EMSLICE":
						value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol))
					case "EM":
						theTable:=olusturulanConfigBu.getAltTableByName(olusturulanConfigBu.AltTable[lol])
						for d := 0; d < reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Len(); d++ {
							switch theTable.CinsAnalysis[d]{
							case "string":
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
							case "int":
								value.Field(i).Index(turn).Field(lol).Field(d).SetInt(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(int64))
							case "bool":
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
							case "float","float32","float64":
								value.Field(i).Index(turn).Field(lol).Field(d).SetFloat(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(float64))
							default:
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))

							}
							/*	if d==2{
									value.Field(i).Index(turn).Field(lol).Field(d).SetInt(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(int64))
								}else{
									value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
								}*/
						}
						//}
						//fmt.Println(value.Field(i))
						//fmt.Println(turn)
						//fmt.Println(lol)

						//  fmt.Println(gelen.Index(lol))

						//fmt.Println(gelen.Index(lol).Interface())
						//fmt.Println(reflect.ValueOf(gelen.Index(lol).Interface()).Len())
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Len())
						//for columnCounter:=0;columnCounter<reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Index(0).Interface()).Len();columnCounter++ {

						//fmt.Println(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Index(0))
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Len())
						//fmt.Println(value.Field(i).Index(turn).Field(lol).Field(2))
						//fmt.Println(value.Field(i).Index(turn).Field(lol).Type())
						//fmt.Println(gelen.Index(lol).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol).Index(turn)) eskisi buydu
					default:

					}
				}

			}
			//value.Field(i).Set(reflect.ValueOf(olusturulanConfigBu.readStructFromDB(Id,Mssql,reflect.New(underlyingValue).Elem().Interface(),true,recNumber+1)))
		case "EMSLICE":
			underlyingValue:=value.Field(i).Type().Elem()
			olusturulanConfigBu := createOrmConfigForMySQl(reflect.New(underlyingValue).Elem().Interface(),config.TableName+"_"+strconv.Itoa(i))
			countOfTuples := getCountOfTuplesMySQl(Id,mysql,olusturulanConfigBu.TableName)

			gelen:=reflect.ValueOf(olusturulanConfigBu.readStructFromDB(Id,mysql,reflect.New(underlyingValue).Elem().Interface(),true, countOfTuples)) //it returns a slice, each element for a column
			//fmt.Println(gelen)
			value.Field(i).Set(reflect.MakeSlice(value.Field(i).Type(),countOfTuples,countOfTuples))
			for turn:=0;turn<countOfTuples;turn++{

				for lol:=0;lol<gelen.Len();lol++{
					switch olusturulanConfigBu.TypeAnalysis[lol] {
					case "NORMAL":

						value.Field(i).Index(turn).Field(lol).Set(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol).Index(turn))
					case "ARR":
						for lollol:=0;lollol<gelen.Index(lol).Len();lollol++{
							value.Field(i).Index(turn).Field(lol).Index(lollol).Set(gelen.Index(lol).Index(lollol))
						}
						/*for ii,vv:=range v{
							value.Field(i).Index(turn).Field(lol).Index(ii).Set(vv)
						}*/
					case "SLICE":
						value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol))
					case "EMARR":
						for lollol:=0;lollol<gelen.Index(lol).Len();lollol++{
							value.Field(i).Index(turn).Field(lol).Index(lollol).Set(gelen.Index(lol).Index(lollol))
						}
						/*for ii,vv:=range v{
							value.Field(i).Index(turn).Field(lol).Index(ii).Set(vv)
						}*/
					case "EMSLICE":
						value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol))
					case "EM":
						theTable:=olusturulanConfigBu.getAltTableByName(olusturulanConfigBu.AltTable[lol])
						for d := 0; d < reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Len(); d++ {
							switch theTable.CinsAnalysis[d]{
							case "string":
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
							case "int":
								value.Field(i).Index(turn).Field(lol).Field(d).SetInt(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(int64))
							case "bool":
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
							case "float","float32","float64":
								value.Field(i).Index(turn).Field(lol).Field(d).SetFloat(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(float64))
							default:
								value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))

							}
							/*	if d==2{
									value.Field(i).Index(turn).Field(lol).Field(d).SetInt(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn).Interface().(int64))
								}else{
									value.Field(i).Index(turn).Field(lol).Field(d).Set(reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Index(d).Interface()).Index(turn))
								}*/
						}
						//}
						//fmt.Println(value.Field(i))
						//fmt.Println(turn)
						//fmt.Println(lol)

						//  fmt.Println(gelen.Index(lol))

						//fmt.Println(gelen.Index(lol).Interface())
						//fmt.Println(reflect.ValueOf(gelen.Index(lol).Interface()).Len())
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(0).Interface()).Len())
						//for columnCounter:=0;columnCounter<reflect.ValueOf(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Index(0).Interface()).Len();columnCounter++ {

						//fmt.Println(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Index(0))
						//fmt.Println(reflect.ValueOf(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn).Interface()).Len())
						//fmt.Println(value.Field(i).Index(turn).Field(lol).Field(2))
						//fmt.Println(value.Field(i).Index(turn).Field(lol).Type())
						//fmt.Println(gelen.Index(lol).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(reflect.ValueOf(gelen.Index(lol).Interface()).Index(turn))
						//value.Field(i).Index(turn).Field(lol).Set(gelen.Index(lol).Index(turn)) eskisi buydu
					default:

					}
				}

			}
			//value.Field(i).Set(reflect.ValueOf(olusturulanConfigBu.readStructFromDB(Id,Mssql,reflect.New(underlyingValue).Elem().Interface(),true,recNumber+1)))
		case "EM":
			currentOrmConfigForEMTYPE := config.getAltTableByName(config.TableName+"_"+strconv.Itoa(i))
			if isRecursiveCall{
				var objects []interface{}

				theObj := currentOrmConfigForEMTYPE.readStructFromDB(Id,mysql,value.Field(i).Interface(),true,recNumber)
				objects=append(objects,theObj)

				allOfThem=append(allOfThem,objects)
			}else{
				value.Field(i).Set(reflect.ValueOf(currentOrmConfigForEMTYPE.readStructFromDB(Id,mysql,value.Field(i).Interface(),false,0)))
			}


		}
	}
	if isRecursiveCall{
		return allOfThem
	}else{
		return value.Interface()
	}

}

func(config MySqlOrmConfig) getAllAltTables() []MySqlOrmConfig{
	var allOrmConfigs []MySqlOrmConfig
	var waste []MySqlOrmConfig
	var waste2 []MySqlOrmConfig

	waste=append(waste,config)
	waste=append(waste,config.AltTableObj...)
	for _,v:=range waste{
		waste2=append(waste2,v)
		waste2=append(waste2,v.AltTableObj...)
	}

	var empty []MySqlOrmConfig
	val:=false
	for _,v:=range waste2{
		val=false
		for _,k:=range empty{
			if reflect.DeepEqual(k,v){
				val=true
			}
		}

		if !val{
			empty=append(empty,v)
			allOrmConfigs=append(allOrmConfigs,v)
		}
		val=false
	}
	return allOrmConfigs
}

func(config MySqlOrmConfig) getARRtablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMARRTablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMARRTablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="EMARR"{
			result=append(result,config.ContainsEMARRTablesName[i])
		}
	}
	return  result
}

func(config MySqlOrmConfig) getSLItablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMSLICETablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMSLICETablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="EMSLICE"{
			result=append(result,config.ContainsEMSLICETablesName[i])
		}
	}
	return result
}

func(config MySqlOrmConfig) getEMSLItablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMSLICETablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMSLICETablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="SLICE"{
			result=append(result,config.ContainsEMSLICETablesName[i])
		}
	}
	return result
}

func(config MySqlOrmConfig) getEMARRtablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMARRTablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMARRTablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="ARR"{
			result=append(result,config.ContainsEMARRTablesName[i])
		}
	}
	return  result
}
func createConfigAndRelationsReferancesForDatabaseMySql(mssql *MysqlDB,config MySqlOrmConfig){
	//config var ama boş dolduracağız
	allConfigs := config.getAllAltTables()
	for _,k:=range allConfigs{
		for i:=0;i<len(k.ColumnNames);i++{
			var look int = 0
			row:=mssql.Database.QueryRow("Select Count(*) FROM OrcaConfig WHERE TableName=? AND ColumnName=? AND ColumnType=? AND ColumnType2=? AND ColumnOrder=?",k.TableName,k.ColumnNames[i],k.TypeAnalysis[i],k.CinsAnalysis[i],k.ColumnOrder[k.ColumnNames[i]])
			err:=row.Scan(&look)
			if err!=nil{
				look=0
			}
			if look==0{
				_,err=mssql.Database.Exec("Insert into OrcaConfig (TableName,ColumnName,ColumnType,ColumnType2,ColumnOrder) VALUES (?,?,?,?,?)",k.TableName,k.ColumnNames[i],k.TypeAnalysis[i],k.CinsAnalysis[i],k.ColumnOrder[k.ColumnNames[i]])
				Check(err)
			}
		}

		for _,s:=range k.ContainsEMTablesName{
			var look int =0
			row:=mssql.Database.QueryRow("Select Count(*) FROM OrcaRelations WHERE Parent=? AND Child=? AND Slice=? AND Array=? AND Embedded=?",k.TableName,s,false,false,true)
			err:=row.Scan(&look)
			if err!=nil{
				look=0
			}
			if look==0{
				_,err=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (?,?,?,?,?)",k.TableName,s,false,false,true)
				Check(err)
			}
		}

		for _,s:=range k.ContainsEMARRTablesName{
			var look int=0
			row:=mssql.Database.QueryRow("Select Count(*) FROM OrcaRelations WHERE Parent=? AND Child=? AND Slice=? AND Array=? AND Embedded=?",k.TableName,s,false,true,false)
			err:=row.Scan(&look)
			if err!=nil{
				look=0
			}
			if look==0{
				_,err=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (?,?,?,?,?)",k.TableName,s,false,true,false)
				Check(err)
			}
		}

		for _,s:=range k.ContainsEMSLICETablesName{
			var look int=0
			row:=mssql.Database.QueryRow("Select Count(*) FROM OrcaRelations WHERE Parent=? AND Child=? AND Slice=? AND Array=? AND Embedded=?",k.TableName,s,true,false,false)
			err:=row.Scan(&look)
			if err!=nil{
				look=0
			}
			if look==0{
				_,err=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (?,?,?,?,?)",k.TableName,s,true,false,false)
				Check(err)
			}

		}

	}



}

func generateTableForARRorSLIMySql(compare map[string]string,tablename, datatype string,source map[string]reflect.Value,mysql *MysqlDB)(string,[]string){
	var query string ="CREATE TABLE IF NOT EXISTS "+tablename+"(Id int(11) NOT NULL auto_increment, ForeignID int(11),"

	switch datatype {
	case "string","int","float32","float","float64","bool":
		query+="Value "+compare[datatype]+" NULL,"
	default:
		//it should be array or slice of struct
		value := source[tablename]
		underlyingValue:=value.Type().Elem()
		getTheormConfig := createOrmConfigForMySQl(reflect.New(underlyingValue).Elem().Interface(),tablename)
		textQueries :=getTheormConfig.toSQL(mysql)
		createConfigAndRelationsReferancesForDatabaseMySql(mysql,getTheormConfig)
		return "",textQueries
	}


	query+="PRIMARY KEY(Id));"

	return query,nil
}

func getSQLforARRorSLItoADDMySQl(tablename string) string{
	return "INSERT INTO "+tablename+" (ForeignID,Value) VALUES (?,?)"
}

func(config MySqlOrmConfig) getColumnNamesInOrder(toHead, toTail string) string{
	var result string
	for i:=0;i< len(config.ColumnNames);i++{
		result+=toHead+config.ColumnNames[i]+toTail
		if i!= len(config.ColumnNames)-1{
			result+=","
		}
	}
	return result
}

func(config MySqlOrmConfig) getValuesForSQLStringInMySQl() string{
	var result string
	for i:=0;i< len(config.ColumnNames);i++{
		result+="?"
		if i!= len(config.ColumnNames)-1{
			result+=","
		}
	}
	return result
}

func(config MySqlOrmConfig) generateSQLADD(withForeignID bool) string{
	var query string
	if withForeignID{
		query="INSERT INTO "+config.TableName+" (ForeignID,"+config.getColumnNamesInOrder("","")+") VALUES (?,"+config.getValuesForSQLStringInMySQl()+")"
	}else{
		query="INSERT INTO "+config.TableName+" ("+config.getColumnNamesInOrder("","")+") VALUES ("+config.getValuesForSQLStringInMySQl()+")"
	}
	return query
}


func(config MySqlOrmConfig) getAllTableNamesOptimized(Mysql *MysqlDB) []string{

	var allTables []string
	var result []string

	rows,err:=Mysql.Database.Query("Select Parent,Child FROM OrcaRelations")
	Check(err)
	defer rows.Close()
	for rows.Next(){
		var str string
		var str2 string
		rows.Scan(&str,&str2)
		allTables=append(allTables,str,str2)
	}


	for i:=0;i< len(allTables);i++{
		if strings.Contains(allTables[i],"_"){
			fixed:=strings.Split(allTables[i],"_")[0]
			if fixed==config.TableName{
				result=append(result,allTables[i])
			}
		}
	}
	result=append(result,config.TableName)


	return unique(result)
}

func(config MySqlOrmConfig) sqlAdd(mysql *sql.Tx, x interface{},isRecursiveCall bool,Id int){

	var args []interface{}
	args=append(args,Id)
	value:=reflect.ValueOf(x)
	for i:=0;i<value.NumField();i++{
		typeOfvalue := config.TypeAnalysis[i]
		switch typeOfvalue {
		case "NORMAL":
			args=append(args,value.Field(i).Interface())
		case "SLICE","ARR":
			args=append(args,Id)
			for ii:=0;ii<value.Field(i).Len();ii++{
				mysql.Exec(getSQLforARRorSLItoADDMySQl(config.TableName+"_"+strconv.Itoa(i)),Id,value.Field(i).Index(ii).Interface())
			}
		case "EMARR","EMSLICE":
			args=append(args,Id)
			/*	underlyingValue :=value.Field(i).Type().Elem()
				ormConfigFortheEMTYPE:=createOrmConfigforMySQL(reflect.New(underlyingValue).Elem().Interface(),config.TableName+"_"+strconv.Itoa(i))//alttaki işi fora sokup yapacaksın işte
				ormConfigFortheEMTYPE.sqlAdd(mssql,reflect.New(underlyingValue).Elem().Interface(),true)*/
			structOfArrayorSlice:=value.Field(i)
			for s:=0;s<structOfArrayorSlice.Len();s++{
				value.Field(i).Index(s) //bu bir struct
				olusturulanConfigBu:=createOrmConfigForMySQl(value.Field(i).Index(s).Interface(),config.TableName+"_"+strconv.Itoa(i))
				olusturulanConfigBu.sqlAdd(mysql,value.Field(i).Index(s).Interface(),true,Id)
			}
		case "EM":
			args=append(args,Id)
			currentOrmConfigForEMTYPE := config.getAltTableByName(config.TableName+"_"+strconv.Itoa(i))
			currentOrmConfigForEMTYPE.sqlAdd(mysql,value.Field(i).Interface(),true,Id)
		default:

		}
	}
	//if isRecursiveCall{
	_,err:=mysql.Exec(config.generateSQLADD(true),args...)
	Check(err)
	//}else{
	//	mssql.Exec(config.generateSQLADD(false),args...)
	//}
}


func(config MySqlOrmConfig) toSQL(mysql *MysqlDB)[]string{

	var queries []string
	var tables = config.getAllAltTables()
	//just for MSSQL
	var mssqlMap = make(map[string]string)
	mssqlMap["int"]="int(11)"
	mssqlMap["string"]="Text"
	mssqlMap["bool"]="bit"
	mssqlMap["float32"]="float"
	mssqlMap["float"]="float"
	mssqlMap["float64"]="float"
	//

	for t:=0;t< len(tables);t++{
		var query string ="CREATE TABLE IF NOT EXISTS "+tables[t].TableName+" (Id int(11) NOT NULL auto_increment,"
		if t!=0{
			query+="ForeignID int(11) NULL,"
		}
		//definedTypes := []string{"int","bool","float32","float","string"}
		var plus []string
		arrTables :=tables[t].getARRtablesInOrder()
		for arr:=0;arr< len(arrTables);arr++{
			index,_:=strconv.Atoi(strings.Split(arrTables[arr],"_")[1])
			textQuery,_:=generateTableForARRorSLIMySql(mssqlMap,arrTables[arr],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mysql)
			plus = append(plus,textQuery)
		}
		sliTables :=tables[t].getSLItablesInOrder()
		for sli:=0;sli< len(sliTables);sli++{
			index,_:=strconv.Atoi(strings.Split(sliTables[sli],"_")[1])
			textQuery,_:=generateTableForARRorSLIMySql(mssqlMap,sliTables[sli],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mysql)
			plus = append(plus,textQuery)
		}
		emsliTables :=tables[t].getEMSLItablesInOrder()
		for emsli:=0;emsli< len(emsliTables);emsli++{
			index,_:=strconv.Atoi(strings.Split(emsliTables[emsli],"_")[1])
			_,textQueries:=generateTableForARRorSLIMySql(mssqlMap,emsliTables[emsli],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mysql)
			plus = append(plus,textQueries...)
		}
		emarrTables :=tables[t].getEMARRtablesInOrder()
		for emarr:=0;emarr< len(emarrTables);emarr++{
			index,_:=strconv.Atoi(strings.Split(emarrTables[emarr],"_")[1])
			_,textQueries:=generateTableForARRorSLIMySql(mssqlMap,emarrTables[emarr],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mysql)
			plus = append(plus,textQueries...)
		}
		queries=append(queries,plus...)

		for i:=0;i< len(tables[t].TypeAnalysis);i++{

			switch tables[t].TypeAnalysis[i] {
			case "NORMAL": //normal
				query+=buildSQLText(mssqlMap,tables[t].ColumnNames[i],"NORMAL",tables[t].CinsAnalysis[i])
			case "ARR","EMARR"://need another table
				query+=buildSQLText(mssqlMap,tables[t].ColumnNames[i],"ARR",tables[t].CinsAnalysis[i])
			case "SLICE","EMSLICE": //need another table
				query+=buildSQLText(mssqlMap,tables[t].ColumnNames[i],"SLICE",tables[t].CinsAnalysis[i])
			case "EM": //need another table
				query+=buildSQLText(mssqlMap,tables[t].ColumnNames[i],"EM",tables[t].CinsAnalysis[i])
			}
		}

		query+="PRIMARY KEY(Id) );"
		queries=append(queries,query)
	}

	return queries
}

func createOrmConfigForMySQl(x interface{},tableName string) MySqlOrmConfig {

	value := reflect.ValueOf(x)
	var typeAnalysis map[int]string = make(map[int]string)
	var cinsAnalysis map[int]string = make(map[int]string)
	var columnName map[int]string = make(map[int]string)
	var columnOrder map[string]int = make(map[string]int)
	var AltTable map[int]string = make(map[int]string)
	var EmObjects map[string]reflect.Value = make(map[string]reflect.Value)
	var AltTableObj []MySqlOrmConfig
	containsEm,containsEmArr,containsEmSlice,containsUndefined:=false,false,false,false
	containsEmTablesName,containsEmARRTablesName,containsEmSLICETablesName:=[]string{},[]string{},[]string{}

	numfields := value.NumField()
	getTheFieldsNameToMap(value, columnName)
	for i := 0; i < numfields; i++ {
		k := getTheType(value.Field(i).Kind())
		switch k {
		case "int", "float", "bool", "string":
			cinsAnalysis[i] = k
			typeAnalysis[i] = "NORMAL"
			columnOrder[columnName[i]]=i
			AltTable[i]="NULL"
		case "array":
			//array of what
			k2 := getTheType(reflect.ArrayOf(0,value.Field(i).Type()).Kind())
			cinsAnalysis[i] = k2+"_"+reflect.SliceOf(value.Field(i).Type().Elem()).Elem().Name()
			nameOfTheNextTable:=tableName+"_"+strconv.Itoa(i)
			if isNormalArr(reflect.SliceOf(value.Field(i).Type().Elem()).Elem().Name()){
				typeAnalysis[i] = "ARR"
			}else {
				typeAnalysis[i] = "EMARR"
				EmObjects[nameOfTheNextTable]=value.Field(i)
			}
			containsEmArr=true
			//nameOfTheNextTable:=tableName+"_"+columnName[i]+"_EMARR"
			containsEmARRTablesName=append(containsEmARRTablesName,nameOfTheNextTable)
			AltTable[i]=nameOfTheNextTable
			columnOrder[columnName[i]]=i
		case "slice":
			//slice of what
			k2 := getTheType(reflect.SliceOf(value.Field(i).Type()).Kind())
			cinsAnalysis[i] = k2+"_"+reflect.SliceOf(value.Field(i).Type().Elem()).Elem().Name()
			nameOfTheNextTable:=tableName+"_"+strconv.Itoa(i)
			if isNormalSLice(reflect.SliceOf(value.Field(i).Type().Elem()).Elem().Name()){
				typeAnalysis[i] = "SLICE"
			}else{
				typeAnalysis[i] = "EMSLICE"
				EmObjects[nameOfTheNextTable]=value.Field(i)
			}
			containsEmSlice=true
			//nameOfTheNextTable:=tableName+"_"+columnName[i]+"_EMSLICE"

			containsEmSLICETablesName=append(containsEmSLICETablesName,nameOfTheNextTable)
			AltTable[i]=nameOfTheNextTable
			columnOrder[columnName[i]]=i
		case "struct":
			//need another analysis
			cinsAnalysis[i] = k+"_"+value.Field(i).Type().Name()
			typeAnalysis[i] = "EM"
			containsEm=true
			//nameOfTheNextTable:=tableName+"_"+columnName[i]+"_EM"
			nameOfTheNextTable:=tableName+"_"+strconv.Itoa(i)
			containsEmTablesName=append(containsEmTablesName,nameOfTheNextTable)
			columnOrder[columnName[i]]=i
			AltTable[i]=nameOfTheNextTable
			AltTableObj=append(AltTableObj,createOrmConfigForMySQl(value.Field(i).Interface(),nameOfTheNextTable))
		case "undefined":
			//wtf
			cinsAnalysis[i] = "undefined"
			typeAnalysis[i] = "UNDEFINED"
			containsUndefined=true
			columnOrder[columnName[i]]=i
			AltTable[i]="NULL"
		default:
			//omg wtf?!
		}

	}

	return MySqlOrmConfig{
		TableName:                 tableName,
		TypeAnalysis:              typeAnalysis,
		CinsAnalysis:              cinsAnalysis,
		ColumnOrder:               columnOrder,
		ColumnNames:               columnName,
		EMObjects:                 EmObjects,
		AltTable:                  AltTable,
		AltTableObj:               AltTableObj,
		ContainsEM:                containsEm,
		ContainsEMTablesName:      containsEmTablesName,
		ContainsEMARR:             containsEmArr,
		ContainsEMARRTablesName:   containsEmARRTablesName,
		ContainsEMSLICE:           containsEmSlice,
		ContainsEMSLICETablesName: containsEmSLICETablesName,
		ContainsUNDEFINED:         containsUndefined,
	}
}

func (mysql *MysqlDB) AddLocalHooks(hks ...LocalHook) {
	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	mysql.DeleteLocalHooks(ids...)
	mysql.LocalH=append(mysql.LocalH,hks...)
}

func (mysql *MysqlDB) AddLocalHook(hks LocalHook) {
	mysql.DeleteLocalHook(hks.getID())
	mysql.LocalH=append(mysql.LocalH,hks)
}

func (mysql *MysqlDB) DeleteLocalHook(hks string) {
	for i:=0;i< len(mysql.LocalH);i++{
		if mysql.LocalH[i].getID()==hks{
			mysql.LocalH[i]=mysql.LocalH[len(mysql.LocalH)-1]
			mysql.LocalH=mysql.LocalH[:len(mysql.LocalH)-1]
			break
		}
	}
}

func (mysql *MysqlDB) DeleteLocalHooks(hks ...string) {
	mysql.LocalH=reorder(mysql.LocalH,hks)
}

type MysqlOptions struct {
	User string
	Password string
	DatabaseName string
	Network string
}

func(MysqlOption MysqlOptions) Options() []string{
	var ops []string
	ops = append(ops,MysqlOption.User)
	ops = append(ops,MysqlOption.Password)
	ops = append(ops,MysqlOption.DatabaseName)
	ops = append(ops,MysqlOption.Network)

	return ops
}

func getNewIdMySQl(mysql *MysqlDB,tablename string) int{
	currentKey:=0
	row:=mysql.Database.QueryRow("select MAX(Id) from "+tablename)
	row.Scan(&currentKey)
	return currentKey+1
}

func (m *MysqlCollection) Add(x interface{}) {
	if m.lock==false {

		if len(m.LocalH) > 0 {

			//local hook
			var beforeAddLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == BeforeAdd {
					beforeAddLocalHooks = append(beforeAddLocalHooks, m.LocalH[i])
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

	var cacheInterface interface{}

	Id:=getNewIdMySQl(m.Mysql,m.Config.TableName)
	tx,err:=m.Mysql.Database.Begin()
	Check(err)
	m.Config.sqlAdd(tx,x,false,Id)
	err=tx.Commit()
	if err==nil{
		m.ListId=append(m.ListId,Id)
		m.List=append(m.List,x)
		m.CacheList[Id]=cacheInterface
	}
	Check(err)

	if m.lock==false {
		if len(m.LocalH) > 0 {
			//local hook
			var afterAddLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == AfterAdd {
					afterAddLocalHooks = append(afterAddLocalHooks, m.LocalH[i])
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

func (m *MysqlCollection) AddRange(x interface{}) {
	if m.lock==false {

		if len(m.LocalH) > 0 {

			//local hook
			var beforeAddRangeLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == BeforeAddRange {
					beforeAddRangeLocalHooks = append(beforeAddRangeLocalHooks, m.LocalH[i])
				}
			}

			sort.Sort(byPriority(beforeAddRangeLocalHooks))

			for i := 0; i < len(beforeAddRangeLocalHooks); i++ {
				funk := beforeAddRangeLocalHooks[i].getHookFunc()
				x = funk(x)
			}
			//local hook

		}
	}
	var cacheID []int
	var cacheInterface []interface{}
	Id:=getNewIdMySQl(m.Mysql,m.Config.TableName)
	tx,err:=m.Mysql.Database.Begin()
	Check(err)
	value:=reflect.ValueOf(x)
	for i:=0;i<value.Len();i++{
		m.Config.sqlAdd(tx,value.Index(i).Interface(),false,Id)
		cacheID=append(cacheID,i)
		cacheInterface=append(cacheInterface,value.Index(i).Interface())
		Id++
	}
	err=tx.Commit()
	if err==nil&& len(cacheID)!=0{
		for i:=0;i< len(cacheID);i++{
			m.ListId=append(m.ListId,cacheID[i])
			m.List=append(m.List,cacheInterface[i])
			m.CacheList[cacheID[i]]=cacheInterface[i]
		}
	}
	Check(err)

	if m.lock==false {

		if len(m.LocalH) > 0 {

			//local hook
			var afterAddRangeLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == BeforeAddRange {
					afterAddRangeLocalHooks = append(afterAddRangeLocalHooks, m.LocalH[i])
				}
			}

			sort.Sort(byPriority(afterAddRangeLocalHooks))

			for i := 0; i < len(afterAddRangeLocalHooks); i++ {
				funk := afterAddRangeLocalHooks[i].getHookFunc()
				x = funk(x)
			}
			//local hook

		}
	}
}

func (m *MysqlCollection) Update(old interface{}, new interface{}) {

	m.lock=true
	if len(m.LocalH)>0 {
		//local hook
		var beforeUpdateLocalHooks []LocalHook
		for i := 0; i < len(m.LocalH); i++ {
			if _, n := m.LocalH[i].getSign(); n == BeforeUpdate {
				beforeUpdateLocalHooks = append(beforeUpdateLocalHooks, m.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeUpdateLocalHooks))

		for i := 0; i < len(beforeUpdateLocalHooks); i++ {
			funk := beforeUpdateLocalHooks[i].getHookFunc()
			old = funk(old)
		}
		//local hook

	}


	Id:=detectTheIdMySQL(m,old)
	//tables:=m.Config.getAllTablenames()
	tables:=m.Config.getAllTableNamesOptimized(m.Mysql)
	var queries []string
	for i:=0; i<len(tables);i++{
		queries=append(queries,deleteQueryFromTableNameMySQl(tables[i]))
	}

	tx,err:=m.Mysql.Database.Begin()
	Check(err)
	for i:=0;i< len(queries);i++{
		_,err=tx.Exec(queries[i],Id)
	}

	m.Config.sqlAdd(tx,new,false,Id)

	err=tx.Commit()
	if err==nil{
		for i:=0;i< len(m.ListId);i++{
			if m.ListId[i]==Id{
				m.ListId[i]=m.ListId[len(m.ListId)-1]
				m.ListId=m.ListId[:len(m.ListId)-1]
				//
				m.List[i]=m.List[len(m.List)-1]
				m.List=m.List[:len(m.List)-1]
				//
			}
		}


		m.ListId=append(m.ListId,Id)
		m.List=append(m.List,new)
		m.CacheList[Id]=new
	}
	Check(err)






	if len(m.LocalH)>0 {
		//local hook
		var afterUpdateLocalHooks []LocalHook
		for i := 0; i < len(m.LocalH); i++ {
			if _, n := m.LocalH[i].getSign(); n == AfterUpdate {
				afterUpdateLocalHooks = append(afterUpdateLocalHooks, m.LocalH[i])
			}
		}

		sort.Sort(byPriority(afterUpdateLocalHooks))

		for i := 0; i < len(afterUpdateLocalHooks); i++ {
			funk := afterUpdateLocalHooks[i].getHookFunc()
			funk(new)
		}
		//local hook

	}
	m.lock=false


}


func detectTheIdMySQL(m *MysqlCollection,value interface{}) int{
	for i:=0;i< len(m.List);i++{
		if reflect.DeepEqual(m.List[i],value){
			return m.ListId[i]
		}
	}
	return 0
}

func deleteQueryFromTableNameMySQl(tablename string) string{
	return "DELETE FROM "+tablename+" WHERE ForeignID=?"
}

func (m *MysqlCollection) Delete(x interface{}) {
	if m.lock==false {
		if len(m.LocalH) > 0 {
			//local hook
			var beforeDeleteLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == BeforeDelete {
					beforeDeleteLocalHooks = append(beforeDeleteLocalHooks, m.LocalH[i])
				}
			}

			sort.Sort(byPriority(beforeDeleteLocalHooks))

			for i := 0; i < len(beforeDeleteLocalHooks); i++ {
				funk := beforeDeleteLocalHooks[i].getHookFunc()
				x = funk(x)
			}
			//local hook

		}
	}

	Id:=detectTheIdMySQL(m,x)
	//tables:=m.Config.getAllTablenames()
	tables:=m.Config.getAllTableNamesOptimized(m.Mysql)
	var queries []string
	for i:=0; i<len(tables);i++{
		queries=append(queries,deleteQueryFromTableNameMySQl(tables[i]))
	}

	tx,err:=m.Mysql.Database.Begin()
	Check(err)
	for i:=0;i< len(queries);i++{
		_,err=tx.Exec(queries[i],Id)
	}
	err=tx.Commit()
	if err==nil{
		for i:=0;i< len(m.ListId);i++{
			if i==Id{
				m.ListId[i]=m.ListId[len(m.ListId)-1]
				m.ListId=m.ListId[:len(m.ListId)-1]
				//
				m.List[i]=m.List[len(m.List)-1]
				m.List=m.List[:len(m.List)-1]
				//
			}
		}
		delete(m.CacheList,Id)
	}


	if m.lock==false {
		if len(m.LocalH) > 0 {

			//local hook
			var afterDeleteLocalHooks []LocalHook
			for i := 0; i < len(m.LocalH); i++ {
				if _, n := m.LocalH[i].getSign(); n == AfterDelete {
					afterDeleteLocalHooks = append(afterDeleteLocalHooks, m.LocalH[i])
				}
			}

			sort.Sort(byPriority(afterDeleteLocalHooks))

			for i := 0; i < len(afterDeleteLocalHooks); i++ {
				funk := afterDeleteLocalHooks[i].getHookFunc()
				funk(x)
			}
			//local hook

		}
	}
}

func (m *MysqlCollection) Clear() {

	tx,err:=m.Mysql.Database.Begin()
	Check(err)
	tables:=m.Config.getAllTableNamesOptimized(m.Mysql)
	for i:=0;i< len(tables);i++{
		_,err:=tx.Exec("TRUNCATE TABLE "+tables[i])
		Check(err)
	}
	err=tx.Commit()
	if err==nil{
		m.List=m.List[:0]
		m.ListId=m.ListId[:0]
		for k,_:= range m.CacheList{
			delete(m.CacheList,k)
		}
	}
	Check(err)
}

func (m *MysqlCollection) Foreach(i interface{}) {
	panic("implement me")
}

func (m *MysqlCollection) GetLogs() {
	panic("implement me")
}

func (m *MysqlCollection) ToSlice() []interface{} {
	slice:=make([]interface{},0,1)

	for _,v:=range m.List{
		slice=append(slice,v)
	}

	return slice
}
