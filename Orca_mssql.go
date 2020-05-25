package Orca

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"reflect"
	"strconv"
	"strings"
)

type MssqlDB struct {
	Name string
	Database *sql.DB
	Timeout string
	ConnectionString string
	LocalH []LocalHook


}

func getMssqlDB(dbName,connectionString,timeOut string) *MssqlDB{
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;", "Orca\\SQL2017", "sa", "sasa", 1433)
	db,err:=sql.Open("sqlserver",connString)
	Check(err)

	//check database is exists
	_,err=db.Exec("IF NOT EXISTS (SELECT name FROM master.sys.databases WHERE name = N'"+dbName+"') BEGIN Create database "+dbName+"; END;")
	Check(err)

	db,err=sql.Open("sqlserver",connString+"database="+dbName+";")
    Check(err)


	return &MssqlDB{
		Name:dbName,
		Database:db,
		Timeout: timeOut,
		ConnectionString:connectionString,
		LocalH: []LocalHook{},
	}
}

func(mssql *MssqlDB) AddLocalHooks(hks ...LocalHook){
	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	mssql.DeleteLocalHooks(ids...)
	mssql.LocalH=append(mssql.LocalH,hks...)
}
func(mssql *MssqlDB) AddLocalHook(hks LocalHook){
	mssql.DeleteLocalHook(hks.getID())
	mssql.LocalH=append(mssql.LocalH,hks)
}
func(mssql *MssqlDB) DeleteLocalHook(hks string){
	for i:=0;i< len(mssql.LocalH);i++{
		if mssql.LocalH[i].getID()==hks{
			mssql.LocalH[i]=mssql.LocalH[len(mssql.LocalH)-1]
			mssql.LocalH=mssql.LocalH[:len(mssql.LocalH)-1]
			break
		}
	}
}
func(mssql *MssqlDB) DeleteLocalHooks(hks ...string){

	mssql.LocalH=reorder(mssql.LocalH,hks)
}

type MssqlCollection struct {
	Mssql *MssqlDB
	ListId []int
	List []interface{}
	LocalH []LocalHook
	Timeout string
	ConnectionString string
	DatabaseName string
	Config ormConfig
	CacheList map[int]interface{}
}

type MssqlOptions struct {
	ConnectionString string
	Timeout string
	DbName string
}

func(MsOptions MssqlOptions) Options() []string{
	var ops []string
	ops = append(ops,MsOptions.ConnectionString)
	ops = append(ops,MsOptions.Timeout)
	ops = append(ops,MsOptions.DbName)

	return ops
}

func SetMSSQLOptions(connectionString , timeout, dbName string) IOptions{
	return MssqlOptions{
		ConnectionString: connectionString,
		Timeout:          timeout,
		DbName:dbName,
	}
}

func getTheType(kind reflect.Kind) string{

	switch kind{
	case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Uint,reflect.Uint8,reflect.Uint32,reflect.Uint16,reflect.Uint64:
		return "int"
	case reflect.Float32,reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Slice:
		return "slice"
	case reflect.Array:
        return "array"
	case reflect.Struct:
		return "struct"
	default:
		return "undefined"
	}
}
func getTheFieldsName(value reflect.Value)  (names []string) {
	for i:=0;i<value.NumField();i++{
		names = append(names,value.Type().Field(i).Name)
	}
	return
}

func getTheFieldsNameToMap(value reflect.Value,theMap map[int]string){
	for i:=0;i<value.NumField();i++{
		theMap[i]=value.Type().Field(i).Name
	}
}

type ormRelations struct {
	ParentTableName string
	ChildTableName []string
	IsSlice []bool
	IsArray []bool
	IsEmbedded []bool
}

type ormConfig struct {
	TableName string
	TypeAnalysis map[int]string
	CinsAnalysis map[int]string
	ColumnOrder map[string]int
	ColumnNames map[int]string
	AltTable map[int]string
	EMObjects map[string]reflect.Value
	AltTableObj []ormConfig
	ContainsEM bool
	ContainsEMTablesName []string
	ContainsEMARR bool
	ContainsEMARRTablesName []string
	ContainsEMSLICE bool
	ContainsEMSLICETablesName []string
	ContainsUNDEFINED bool
}

func(config ormConfig) columnCount() int{
	return len(config.ColumnOrder)
}

func(config ormConfig) hasAltTable() bool{
	return config.AltTableObj!=nil
}

func(config ormConfig) altTableCount() int{
	return len(config.AltTableObj)
}

func(config ormConfig) isSame(config2 ormConfig) bool{ //this function does not check the alttable objects

	var c1 = config.TableName==config2.TableName
	if !c1{
		return false
	}
	//var c2 bool
	if len(config.TypeAnalysis)!= len(config2.TypeAnalysis){
		return false
	}
	for i:=0;i< len(config.TypeAnalysis);i++{
      if config.TypeAnalysis[i]!=config2.TypeAnalysis[i]{
		  return false
	  }
	}

	if len(config.CinsAnalysis)!= len(config2.CinsAnalysis){
		return false
	}

	for i:=0;i< len(config.CinsAnalysis);i++{
		if config.CinsAnalysis[i]!=config2.CinsAnalysis[i]{
			return false
		}
	}

	if len(config.ColumnNames)!= len(config2.ColumnNames){
		return false
	}
	var names []string
	for i:=0;i< len(config.ColumnNames);i++{
		if config.ColumnNames[i]!=config2.ColumnNames[i]{
			return false
		}
		names=append(names,config.ColumnNames[i])
	}

	if len(config.ColumnOrder)!= len(config2.ColumnOrder){
		return false
	}

	if len(config.ColumnOrder)!= len(names){
		return false
	}

	for i:=0;i< len(names);i++{
		if config.ColumnOrder[names[i]]!=config2.ColumnOrder[names[i]]{
			return false
		}
	}

	if config.ContainsEM!=config2.ContainsEM{
		return false
	}

	if config.ContainsEMARR!=config2.ContainsEMARR{
		return false
	}

	if config.ContainsEMSLICE!=config2.ContainsEMSLICE{
		return false
	}

	if config.ContainsUNDEFINED!=config2.ContainsUNDEFINED{
		return false
	}

	if len(config.ContainsEMTablesName)!= len(config2.ContainsEMTablesName){
		return false
	}
	for i:=0;i< len(config.ContainsEMTablesName);i++{
		if config.ContainsEMTablesName[i]!=config2.ContainsEMTablesName[i]{
			return false
		}
	}

	if len(config.ContainsEMARRTablesName)!= len(config2.ContainsEMARRTablesName){
		return false
	}

	for i:=0;i< len(config.ContainsEMARRTablesName);i++{
		if config.ContainsEMARRTablesName[i]!=config2.ContainsEMARRTablesName[i]{
			return false
		}
	}

	if len(config.ContainsEMSLICETablesName)!= len(config2.ContainsEMSLICETablesName){
		return false
	}

	for i:=0;i< len(config.ContainsEMSLICETablesName);i++{
		if config.ContainsEMSLICETablesName[i]!=config2.ContainsEMSLICETablesName[i]{
			return false
		}
	}

	if len(config.AltTableObj)!= len(config2.AltTableObj){
		return false
	}

	for i:=0;i< len(config.AltTableObj);i++{
		if !config.AltTableObj[i].isSame(config2.AltTableObj[i]){
			return false
		}
	}

	return true
}

func(config ormConfig) getAllAltTables() []ormConfig{
	var allOrmConfigs []ormConfig
	var waste []ormConfig
	var waste2 []ormConfig

	waste=append(waste,config)
	waste=append(waste,config.AltTableObj...)
	for _,v:=range waste{
		waste2=append(waste2,v)
        waste2=append(waste2,v.AltTableObj...)
	}

	var empty []ormConfig
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

func(config ormConfig) getARRtablesInOrder() []string{
	var result []string
   for i:=0;i< len(config.ContainsEMARRTablesName);i++{
   	 index,_:=strconv.Atoi(strings.Split(config.ContainsEMARRTablesName[i],"_")[1])
   	 if config.TypeAnalysis[index]!="EMARR"{
   	 	result=append(result,config.ContainsEMARRTablesName[i])
	 }
   }
   return  result
}
func(config ormConfig) getSLItablesInOrder() []string{
 var result []string
 for i:=0;i< len(config.ContainsEMSLICETablesName);i++{
	 index,_:=strconv.Atoi(strings.Split(config.ContainsEMSLICETablesName[i],"_")[1])
	 if config.TypeAnalysis[index]!="EMSLICE"{
		 result=append(result,config.ContainsEMSLICETablesName[i])
	 }
 }
 return result
}
func(config ormConfig) getEMARRtablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMARRTablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMARRTablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="ARR"{
			result=append(result,config.ContainsEMARRTablesName[i])
		}
	}
	return  result
}
func(config ormConfig) getEMSLItablesInOrder() []string{
	var result []string
	for i:=0;i< len(config.ContainsEMSLICETablesName);i++{
		index,_:=strconv.Atoi(strings.Split(config.ContainsEMSLICETablesName[i],"_")[1])
		if config.TypeAnalysis[index]!="SLICE"{
			result=append(result,config.ContainsEMSLICETablesName[i])
		}
	}
	return result
}
func(config ormConfig) generateSQLADD(withForeignID bool) string{
	var query string
	if withForeignID{
		query="INSERT INTO "+config.TableName+" (ForeignID,"+config.getColumnNamesInOrder("","")+") VALUES (@ForeignID,"+config.getColumnNamesInOrder("@","")+")"
	}else{
		query="INSERT INTO "+config.TableName+" ("+config.getColumnNamesInOrder("","")+") VALUES ("+config.getColumnNamesInOrder("@","")+")"
	}
     return query
}
func(config ormConfig) generateSQLUPDATE(IndexOfnewValues []int) string{
	stringList:=strings.Split(config.getColumnNamesInOrder("","="),",")
	query:="UPDATE "+config.TableName+" SET "
	for i:=0;i< len(stringList);i++{
		for k:=0;k< len(IndexOfnewValues);k++{
			if i==IndexOfnewValues[k]{
				query+=stringList[i]+"@"+config.ColumnNames[i]
				if i!= len(config.ColumnNames)-1{
					query+=", "
				}
			}
		}
	}
	query+=" WHERE "
	for i:=0;i< len(stringList);i++{
		query+=stringList[i]+"@"+config.ColumnNames[i]
		if i!= len(config.ColumnNames)-1{
			query+=" AND "
		}
	}
	return query
}
func(config ormConfig) generateSQLDELETE() string{
	stringList:=strings.Split(config.getColumnNamesInOrder("","="),",")
	query:="DELETE FROM "+config.TableName+" WHERE "
	for i:=0;i< len(config.ColumnNames);i++{
		query+=stringList[i]+"@"+config.ColumnNames[i]
		if i!= len(config.ColumnNames)-1{
			query+=" AND "
		}
	}
	return query
}
func(config ormConfig) generateSQLCLEAR() string{
	query:="DELETE FROM "+config.TableName
	return query
}
func(config ormConfig) getColumnNamesInOrder(toHead, toTail string) string{
	var result string
	for i:=0;i< len(config.ColumnNames);i++{
		result+=toHead+config.ColumnNames[i]+toTail
		if i!= len(config.ColumnNames)-1{
			result+=","
		}
	}
	return result
}

func(config ormConfig) getAltTableByName(name string) ormConfig{
	dz:=config.getAllAltTables()
	for i:=0;i< len(dz);i++{
		if dz[i].TableName==name{
			return dz[i]
		}
	}
	return ormConfig{}
}

func(config ormConfig) sqlAdd(mssql *sql.Tx,x interface{},isRecursiveCall bool,Id int){

	var args []interface{}
	value:=reflect.ValueOf(x)
	for i:=0;i<value.NumField();i++{
		typeOfvalue := config.TypeAnalysis[i]
		switch typeOfvalue {
		case "NORMAL":
			args=append(args,sql.Named(config.ColumnNames[i],value.Field(i).Interface()))
		case "SLICE","ARR":
			args=append(args,sql.Named(config.ColumnNames[i],Id))
            for ii:=0;ii<value.Field(i).Len();ii++{
            	mssql.Exec(getSQLforARRorSLItoADD(config.TableName+"_"+strconv.Itoa(i)),sql.Named("ForeignID",Id),sql.Named("Value",value.Field(i).Index(ii).Interface()))
			}
		case "EMARR","EMSLICE":
			args=append(args,sql.Named(config.ColumnNames[i],Id))
		/*	underlyingValue :=value.Field(i).Type().Elem()
			ormConfigFortheEMTYPE:=createOrmConfig(reflect.New(underlyingValue).Elem().Interface(),config.TableName+"_"+strconv.Itoa(i))//alttaki işi fora sokup yapacaksın işte
			ormConfigFortheEMTYPE.sqlAdd(mssql,reflect.New(underlyingValue).Elem().Interface(),true)*/
		 structOfArrayorSlice:=value.Field(i)
		 for s:=0;s<structOfArrayorSlice.Len();s++{
		 	value.Field(i).Index(s) //bu bir struct
		 	olusturulanConfigBu:=createOrmConfig(value.Field(i).Index(s).Interface(),config.TableName+"_"+strconv.Itoa(i))
		 	olusturulanConfigBu.sqlAdd(mssql,value.Field(i).Index(s).Interface(),true,Id)
		 }
		case "EM":
			args=append(args,sql.Named(config.ColumnNames[i],Id))
            currentOrmConfigForEMTYPE := config.getAltTableByName(config.TableName+"_"+strconv.Itoa(i))
			currentOrmConfigForEMTYPE.sqlAdd(mssql,value.Field(i).Interface(),true,Id)
		default:

		}
	}
	if isRecursiveCall{
		args=append(args,sql.Named("ForeignID",Id))
		mssql.Exec(config.generateSQLADD(true),args...)
	}else{
		mssql.Exec(config.generateSQLADD(false),args...)
	}
}

func getNewId(mssql *MssqlDB,tablename string) int{
	currentKey:=0
	row:=mssql.Database.QueryRow("select MAX(Id) from "+tablename)
	row.Scan(&currentKey)
	return currentKey+1
}

func getSQLforARRorSLItoADD(tablename string) string{
	return "INSERT INTO "+tablename+" (ForeignID,Value) VALUES (@ForeignID,@Value)"
}

func buildSQLText(compare map[string]string,columnname string,t,c string) string{
    var returnTheString string=""
	switch t {
	case "NORMAL":
		returnTheString+=columnname+" "+compare[c]+" NULL,"
	case "ARR"://need fk
		returnTheString+=columnname+" "+compare["int"]+" NULL,"
	case "SLICE": //need fk
		returnTheString+=columnname+" "+compare["int"]+" NULL,"
	case "EM": //need fk
		returnTheString+=columnname+" "+compare["int"]+" NULL,"
	}

	return returnTheString
}

func generateTableForARRorSLI(compare map[string]string,tablename, datatype string,source map[string]reflect.Value,mssql *MssqlDB) (string,[]string){
	var query string ="IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = N'"+tablename+"') BEGIN CREATE TABLE [dbo].["+tablename+"]([Id] [int] IDENTITY(1,1) NOT NULL, ForeignID [int],"

	switch datatype {
	case "string","int","float32","float","float64","bool":
		query+="Value "+compare[datatype]+" NULL,"
	default:
		//it should be array or slice of struct
		value := source[tablename]
        underlyingValue:=value.Type().Elem()
		getTheormConfig := createOrmConfig(reflect.New(underlyingValue).Elem().Interface(),tablename)
		textQueries :=getTheormConfig.toSQL(mssql)
		createConfigAndRelationsReferancesForDatabase(mssql,getTheormConfig)
		return "",textQueries
	}


	query+="CONSTRAINT [PK_"+tablename+"] PRIMARY KEY CLUSTERED([Id] ASC)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]) ON [PRIMARY]; END;"

	return query,nil
}

func(config ormConfig) toSQL(mssql *MssqlDB) []string{

	var queries []string
	var tables = config.getAllAltTables()
	//just for MSSQL
	var mssqlMap = make(map[string]string)
	mssqlMap["int"]="[int]"
	mssqlMap["string"]="[nvarchar](max)"
	mssqlMap["bool"]="[BIT]"
	mssqlMap["float32"]="[FLOAT]"
	mssqlMap["float"]="[FLOAT]"
	mssqlMap["float64"]="[FLOAT]"
	//

	for t:=0;t< len(tables);t++{
	var query string ="IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = N'"+tables[t].TableName+"') BEGIN CREATE TABLE [dbo].["+tables[t].TableName+"]([Id] [int] IDENTITY(1,1) NOT NULL,"
	if t!=0{
		query+="ForeignID [int] NULL,"
	}
	//definedTypes := []string{"int","bool","float32","float","string"}
		var plus []string
		arrTables :=tables[t].getARRtablesInOrder()
		for arr:=0;arr< len(arrTables);arr++{
			index,_:=strconv.Atoi(strings.Split(arrTables[arr],"_")[1])
			textQuery,_:=generateTableForARRorSLI(mssqlMap,arrTables[arr],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mssql)
			plus = append(plus,textQuery)
		}
		sliTables :=tables[t].getSLItablesInOrder()
		for sli:=0;sli< len(sliTables);sli++{
			index,_:=strconv.Atoi(strings.Split(sliTables[sli],"_")[1])
			textQuery,_:=generateTableForARRorSLI(mssqlMap,sliTables[sli],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mssql)
			plus = append(plus,textQuery)
		}
		emsliTables :=tables[t].getEMSLItablesInOrder()
		for emsli:=0;emsli< len(emsliTables);emsli++{
			index,_:=strconv.Atoi(strings.Split(emsliTables[emsli],"_")[1])
			_,textQueries:=generateTableForARRorSLI(mssqlMap,emsliTables[emsli],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mssql)
			plus = append(plus,textQueries...)
		}
		emarrTables :=tables[t].getEMARRtablesInOrder()
		for emarr:=0;emarr< len(emarrTables);emarr++{
			index,_:=strconv.Atoi(strings.Split(emarrTables[emarr],"_")[1])
			_,textQueries:=generateTableForARRorSLI(mssqlMap,emarrTables[emarr],strings.Split(tables[t].CinsAnalysis[index],"_")[1],tables[t].EMObjects,mssql)
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

	query+="CONSTRAINT [PK_"+tables[t].TableName+"] PRIMARY KEY CLUSTERED([Id] ASC)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]) ON [PRIMARY]; END;"
	queries=append(queries,query)
	}

    return queries
}

func isNormalArr(t string) bool{
	 result := true
	switch t {
	case "string":
	case "bool":
	case "float":
	case "float32":
	case "float64":
	case "int":
	default:
		result=false
	}
	return result
}

func isNormalSLice(t string) bool{
	result := true
	switch t {
	case "string":
	case "bool":
	case "float":
	case "float32":
	case "float64":
	case "int":
	default:
		result=false
	}
	return result
}

/*func(config ormConfig) getRelation() ormRelations{

	child_Slices := config.ContainsEMSLICETablesName
	child_Arrays :=config.ContainsEMARRTablesName
	child_Structs := config.ContainsEMTablesName
	var names []string
	names=append(names,child_Slices...)
	names=append(names,child_Arrays...)
	names=append(names,child_Structs...)

	return ormRelations{
		ParentTableName: config.TableName,
		ChildTableName:  ,
		IsSlice:         nil,
		IsArray:         nil,
		IsEmbedded:      nil,
	}

}*/

func createOrmConfig(x interface{},tableName string) ormConfig { //must be a struct


	value := reflect.ValueOf(x)
	var typeAnalysis map[int]string = make(map[int]string)
	var cinsAnalysis map[int]string = make(map[int]string)
	var columnName map[int]string = make(map[int]string)
	var columnOrder map[string]int = make(map[string]int)
	var AltTable map[int]string = make(map[int]string)
	var EmObjects map[string]reflect.Value = make(map[string]reflect.Value)
	var AltTableObj []ormConfig
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
			AltTableObj=append(AltTableObj,createOrmConfig(value.Field(i).Interface(),nameOfTheNextTable))
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

	return ormConfig{
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

func createConfigAndRelationsReferancesForDatabase(mssql *MssqlDB,config ormConfig){ //this function create/fill config and relations table on the database


	//config var ama boş dolduracağız
	allConfigs := config.getAllAltTables()
	    for _,k:=range allConfigs{
			for i:=0;i<len(k.ColumnNames);i++{
				_,err:=mssql.Database.Exec("Insert into OrcaConfig (TableName,ColumnName,ColumnType,ColumnType2,ColumnOrder) VALUES (@TableName,@ColumnName,@ColumnType,@ColumnType2,@ColumnOrder)",sql.Named("TableName",k.TableName,),sql.Named("ColumnName",k.ColumnNames[i]),sql.Named("ColumnType",k.TypeAnalysis[i]),sql.Named("ColumnType2",k.CinsAnalysis[i]),sql.Named("ColumnOrder",k.ColumnOrder[k.ColumnNames[i]]))
				Check(err)
			}

			for _,s:=range k.ContainsEMTablesName{
				_,err:=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (@Parent,@Child,@Slice,@Array,@Embedded)",sql.Named("Parent",k.TableName),sql.Named("Child",s),sql.Named("Slice",false),sql.Named("Array",false),sql.Named("Embedded",true))
				Check(err)
			}

			for _,s:=range k.ContainsEMARRTablesName{
				_,err:=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (@Parent,@Child,@Slice,@Array,@Embedded)",sql.Named("Parent",k.TableName),sql.Named("Child",s),sql.Named("Slice",false),sql.Named("Array",true),sql.Named("Embedded",false))
				Check(err)
			}

			for _,s:=range k.ContainsEMSLICETablesName{
				_,err:=mssql.Database.Exec("Insert into OrcaRelations (Parent,Child,Slice,Array,Embedded) VALUES (@Parent,@Child,@Slice,@Array,@Embedded)",sql.Named("Parent",k.TableName),sql.Named("Child",s),sql.Named("Slice",true),sql.Named("Array",false),sql.Named("Embedded",false))
				Check(err)
			}

		}



	//relations yok onu oluşturacağız

}

func readConfigFromDatabase(mssql *MssqlDB,tablename string) ormConfig{
	var expectedResult ormConfig
	expectedResult.TableName=tablename
	expectedResult.CinsAnalysis=make(map[int]string)
	expectedResult.ColumnNames=make(map[int]string)
	expectedResult.TypeAnalysis=make(map[int]string)
	expectedResult.ColumnOrder=make(map[string]int)
	expectedResult.AltTable = make(map[int]string)

	rows,err:=mssql.Database.Query("Select ColumnName,ColumnType,ColumnType2,ColumnOrder from OrcaConfig WHERE TableName=@TableName",sql.Named("TableName",tablename))
	Check(err)
	var columnName,columnType,columnType2 string
	var columnOrder int
	defer rows.Close()
	for rows.Next(){
         err=rows.Scan(&columnName,&columnType,&columnType2,&columnOrder)
         Check(err)
         expectedResult.ColumnNames[columnOrder]=columnName
         expectedResult.TypeAnalysis[columnOrder]=columnType
         expectedResult.CinsAnalysis[columnOrder]=columnType2
         expectedResult.ColumnOrder[columnName]=columnOrder
	}

	rowsForSlices,err:=mssql.Database.Query("Select Child from OrcaRelations WHERE Parent=@Parent and Slice=@Slice",sql.Named("Parent",tablename),sql.Named("Slice",true))
	Check(err)
	var slices []string
	var value1 string
	defer rowsForSlices.Close()
	for rowsForSlices.Next(){
		rowsForSlices.Scan(&value1)
		slices=append(slices,value1)
		or,_:=strconv.Atoi(strings.Split(value1,"_")[1])
		expectedResult.AltTable[or]=value1
	}

	rowsForArrays,err:=mssql.Database.Query("Select Child from OrcaRelations WHERE Parent=@Parent and Array=@Array",sql.Named("Parent",tablename),sql.Named("Array",true))
	Check(err)
    var arrays []string
	var value2 string
	defer rowsForArrays.Close()
	for rowsForArrays.Next(){
		rowsForArrays.Scan(&value2)
		arrays=append(arrays,value2)
		or,_:=strconv.Atoi(strings.Split(value2,"_")[1])
		expectedResult.AltTable[or]=value2
	}

	rowsForStructs,err:=mssql.Database.Query("Select Child from OrcaRelations WHERE Parent=@Parent and Embedded=@Embedded",sql.Named("Parent",tablename),sql.Named("Embedded",true))
    Check(err)
	var structs []string
	var value3 string
	defer rowsForStructs.Close()
	for rowsForStructs.Next(){
		rowsForStructs.Scan(&value3)
		structs=append(structs,value3)
		or,_:=strconv.Atoi(strings.Split(value3,"_")[1])
		expectedResult.AltTable[or]=value3
	}

	if len(slices)>0{
		expectedResult.ContainsEMSLICE=true
		expectedResult.ContainsEMSLICETablesName=slices
	}else{
		expectedResult.ContainsEMSLICE=false
		expectedResult.ContainsEMSLICETablesName=[]string{}
	}

	if len(arrays)>0{
		expectedResult.ContainsEMARR=true
		expectedResult.ContainsEMARRTablesName=arrays
	}else{
		expectedResult.ContainsEMARR=false
		expectedResult.ContainsEMARRTablesName=[]string{}
	}

	if len(structs)>0{
		expectedResult.ContainsEM=true
		expectedResult.ContainsEMTablesName=structs
	}else{
		expectedResult.ContainsEM=false
		expectedResult.ContainsEMTablesName=[]string{}
	}


    if len(expectedResult.AltTable)>0{
    	var orderedChilds []string
    	for i:=0;i< len(expectedResult.CinsAnalysis);i++{
    		if strings.Split(expectedResult.CinsAnalysis[i],"_")[0]=="struct"{
    			orderedChilds=append(orderedChilds,expectedResult.CinsAnalysis[i])
			}
		}

		var ormconfigArray []ormConfig

    	for i:=0;i< len(orderedChilds);i++{
    		var index int
    		for k:=0;k< len(expectedResult.CinsAnalysis);k++{
    			 if expectedResult.CinsAnalysis[k]==orderedChilds[i]{
    			 	index = k
				 }
			}
			ormconfigArray=append(ormconfigArray,readConfigFromDatabase(mssql,tablename+"_"+strconv.Itoa(index)))
		}

		expectedResult.AltTableObj=ormconfigArray
	}

	return expectedResult

}

func executeQueries(mssql *MssqlDB,queries []string){
	for i:=0;i< len(queries);i++{
		mssql.Database.Exec(queries[i])
	}
}

func fixTheQueries(queries []string,mainTable string) []string{
	for i:=0;i< len(queries);i++{
		if !strings.Contains(queries[i],"TABLE_NAME = N'"+mainTable+"')"){
			if !strings.Contains(queries[i],"ForeignID"){
				queries[i]=strings.Replace(queries[i],"([Id] [int] IDENTITY(1,1) NOT NULL,","([Id] [int] IDENTITY(1,1) NOT NULL,ForeignID [int] NULL,",1)
			}
		}
	}
	return  queries
}

func readStructFromDB(Id int,Mssql *MssqlDB,x interface{},config ormConfig) interface{}{

}

func(mssql *MssqlDB) GetCollection(x interface{},tableName string) ICollection{

	/*
	CREATE TABLE [dbo].[denemeconfig]([Id] [int] IDENTITY(1,1) NOT NULL,[TableAd] [nvarchar](255) NULL,[KolonAd] [nvarchar](255) NULL,[KolonTip] [nvarchar](255) NULL,[Cinsi] [nvarchar](50) NULL,[KolonSira] [int] NULL,[KarsilikGelen] [nvarchar](255) NULL,CONSTRAINT [PK_denemeconfig] PRIMARY KEY CLUSTERED([Id] ASC)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]) ON [PRIMARY]
	*/
	//tableconfig varmı kontrol et yoksa oluştur
	mssql.Database.Exec("IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = N'"+"OrcaConfig"+"') BEGIN CREATE TABLE [dbo].[OrcaConfig]([Id] [int] IDENTITY(1,1) NOT NULL,[TableName] [nvarchar](255) NULL,[ColumnName] [nvarchar](255) NULL,[ColumnType] [nvarchar](255) NULL,[ColumnType2] [nvarchar](50) NULL,[ColumnOrder] [int] NULL,CONSTRAINT [PK_denemeconfig] PRIMARY KEY CLUSTERED([Id] ASC)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]) ON [PRIMARY]; END;")
	//ilgili tablo varmı kontrol et yeni mi değilmi bak yoksa oluştur
	_,errr:=mssql.Database.Exec("IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = N'"+"OrcaRelations"+"') BEGIN CREATE TABLE [dbo].[OrcaRelations]([Id] [int] IDENTITY(1,1) NOT NULL,[Parent] [nvarchar](255) NULL,[Child] [nvarchar](255) NULL,[Slice] [BIT] NOT NULL,[Array] [BIT] NOT NULL,[Embedded] [BIT] NOT NULL,PRIMARY KEY CLUSTERED([Id] ASC) WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]) ON [PRIMARY]; END;")
	Check(errr)
	var count int
	row:=mssql.Database.QueryRow("SELECT COUNT(*) FROM OrcaConfig Where TableName=@table",sql.Named("table",tableName))
	err:=row.Scan(&count)
    Check(err)

	getTheDBSchema:=createOrmConfig(x,tableName)

	allSQLQueries := fixTheQueries(getTheDBSchema.toSQL(mssql),tableName)//tamam artık ürettiğin sqlleri execute et
	update:=getTheDBSchema.generateSQLUPDATE([]int{0,1,2,3})
	deletee:=getTheDBSchema.generateSQLDELETE()
	clear:=getTheDBSchema.generateSQLCLEAR()

    executeQueries(mssql,allSQLQueries)
    _=update
    _=deletee
    _=clear




	if count==0{
     createConfigAndRelationsReferancesForDatabase(mssql,getTheDBSchema)
     //tabloları create et


	}else{
		getTheDBSchemaFromDB:=readConfigFromDatabase(mssql,tableName)
      //şema değişikliği var mı kontrol et yoksa devam
		if getTheDBSchema.isSame(getTheDBSchemaFromDB) {
			//sema degisikligi yok
			//verileri oku
		}else{

		}
	}

	var list []interface{}
	var listID []int
	var cacheList map[int]interface{} = make(map[int]interface{})

	return MssqlCollection{
		Mssql:            mssql,
		ListId:           listID,
		List:             list,
		LocalH:           mssql.LocalH,
		Timeout:          mssql.Timeout,
		ConnectionString: mssql.ConnectionString,
		DatabaseName:     mssql.Name,
		Config:           getTheDBSchema,
		CacheList:        cacheList,
	}
}

func (m MssqlCollection) Add(x interface{}) {
	Id:=getNewId(m.Mssql,m.Config.TableName)
	tx,err:=m.Mssql.Database.Begin()
	Check(err)
	m.Config.sqlAdd(tx,x,false,Id)
	err=tx.Commit()
	Check(err)
}

func (m MssqlCollection) AddRange(interface{}) {
	panic("implement me")
}

func (m MssqlCollection) Update(interface{}, interface{}) {
	panic("implement me")
}

func (m MssqlCollection) Delete(interface{}) {
	panic("implement me")
}

func (m MssqlCollection) Clear() {
	panic("implement me")
}

func (m MssqlCollection) Foreach(interface{}) {
	panic("implement me")
}

func (m MssqlCollection) GetLogs() {
	panic("implement me")
}

func (m MssqlCollection) ToSlice() []interface{} {
	panic("implement me")
}






