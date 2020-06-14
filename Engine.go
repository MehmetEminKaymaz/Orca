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
	AddLocalHooks(...LocalHook)
	AddLocalHook(LocalHook)
	DeleteLocalHook(string)
	DeleteLocalHooks(...string)



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
		return getMysqlDB(options.Options()[0],options.Options()[1],options.Options()[2],options.Options()[3]),nil
	case MSSQL:
		return getMssqlDB(options.Options()[2],options.Options()[0],options.Options()[1]),nil
	case SQLite:
		dname:=options.Options()[0]
		path:=options.Options()[1]
		return getDatabase(dname,path),nil
	case PostgreSQL:
		arr:=options.Options()
		return getPostgreSQL(arr[0],arr[1],arr[2],arr[3],arr[4],arr[5]),nil
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


type LocalHook interface {
	getID() string
	getSign() (int , string)
	getHookFunc() func(interface{})interface{}
}

const(//local hooks
	BeforeAdd="BeforeAdd"
	AfterAdd="AfterAdd"
	BeforeDelete="BeforeDelete"
	AfterDelete="AfterDelete"
	BeforeAddRange="BeforeAddRange"
	AfterAddRange="AfterAddRange"
	BeforeUpdate="BeforeUpdate"
	AfterUpdate="AfterUpdate"

)

type NewLocalHookVariable struct {
	ID string
	Name string
	Priority int
	HookFunc func(interface{})interface{}
}
func (l NewLocalHookVariable) getID() string{
	return l.ID
}

func (l NewLocalHookVariable) getSign() (int,string){

	return l.Priority,l.Name

}
func (l NewLocalHookVariable) getHookFunc() func(interface{}) interface{}{

	return l.HookFunc

}




func NewLocalHook(id string,priority int ,localHook string , functionOftheHook func(interface{})interface{}) LocalHook  {//functionofthehook get current record as a parameter and return updated record

	return NewLocalHookVariable{
		ID:id,
		HookFunc:functionOftheHook,
		Name:localHook,
		Priority:priority,
	}

}

func reorder(data []LocalHook, ids []string) []LocalHook {
	n := len(data)
	i := 0
loop:
	for i < n {
		r := data[i]
		for _, id := range ids {
			if id == r.getID() {
				data[i] = data[n-1]
				n--
				continue loop
			}
		}
		i++
	}
	return data[0:n]
}

type byPriority []LocalHook

func (m byPriority) Len() int { return len(m) }

func (m byPriority) Less(i, j int) bool {
	p,_:=m[i].getSign()
	p2,_:=m[j].getSign()
	return p < p2
}
func (m byPriority) Swap(i, j int) {m[i], m[j] = m[j], m[i]}

//local hooks end

/*

type GlobalHook interface {

	BeforeAdd(interface{}) interface{}
	AfterAdd(interface{}) interface{}
	BeforeDelete(interface{}) interface{}
	AfterDelete(interface{})interface{}
	BeforeUpdate(interface{})interface{}
	AfterUpdate(interface{})interface{}
	BeforeAddRange(interface{})interface{}
	AfterAddRange(interface{})interface{}

}

// x.0,x.1,x.2,x.3,x.4,x.5... => ReplaceWith ''

type OrcaLanguage struct{

}

func(lang OrcaLanguage) Run(intermediateLanguage OLTask,item interface{}) interface{}{

	fnum:=intermediateLanguage.TheFieldNum
	val:=reflect.ValueOf(item)

	for i:=0;i< len(intermediateLanguage.Tasks);i++{
		onTask:=intermediateLanguage.Tasks[i]
		switch onTask.TheTask {
		case ReplaceWith:

			switch intermediateLanguage.Value[i].(type) {
			case string:
				val.Field(fnum).SetString(reflect.ValueOf(intermediateLanguage.Value[i]).String())
			case int,int8,int16,int32,int64:
				val.Field(fnum).SetInt(reflect.ValueOf(intermediateLanguage.Value[i]).Int())
			case bool:
				val.Field(fnum).SetBool(reflect.ValueOf(intermediateLanguage.Value[i]).Bool())
			case float32,float64:
				val.Field(fnum).SetFloat(reflect.ValueOf(intermediateLanguage.Value[i]).Float())
			}

		}
	}


	return val

}

const(

	Start2="x"
	Dot="."
	IfEqual="Equal"
	IfNotEqual="IfNotEqual"
	LessThan="LessThan"
	LessThanOrEqual="LessThanOrEqual"
	GreaterThan="GreaterThan"
	GreaterThanOrEqual="GreaterThanOrEqual"
	ReplaceWith="ReplaceWith"
	Variable="{Orca}"
	Add="Add"

)



type Task struct {
	Order int
	TheTask string
}

type OLTask struct {

   TheFieldNum int
   Tasks []Task
   Value []interface{}

}

type Scanner struct {
    ReadedCode OLTask
}

func (sc *Scanner) Read(s string,k ...interface{}) (OLTask,error){
	 str := ""
	 counter:=0
	 orcaCounter:=0

	 var myTasks OLTask

	 fieldassigned:=false


	 waitforField:=false
	 waitforStart:=true
	 waitforReplaceC:=false
	 waitforReplace:=false
	 waitforDot:=false
     waitforDefault:=false
     waitforVariable:=false
     waitforCMD:=false
	 startErr:=waitforDot||waitforField||waitforReplace||waitforReplaceC||waitforDefault||waitforVariable
     dotErr:=waitforReplaceC||waitforReplace||waitforField||waitforStart||waitforDefault||waitforVariable
    _=dotErr

	for _, c := range s {
		str += string(c)

		switch str {

		case Start2:
			if startErr{
				return OLTask{},errors.New("x was expected!")
			}
			waitforStart=false
			waitforDot=true
			str=""
		case Dot:

			waitforDot=false
			waitforField=true
			str=""

		case IfEqual:
			if waitforCMD{
				waitforVariable=true

			}
		case IfNotEqual:
			if waitforCMD{
				waitforVariable=true

			}
		case LessThan:
			if waitforCMD{
				waitforVariable=true

			}
		case LessThanOrEqual:
			if waitforCMD{
				waitforVariable=true

			}
		case GreaterThan:
			if waitforCMD{
				waitforVariable=true

			}
		case GreaterThanOrEqual:
			if waitforCMD{
				waitforVariable=true


			}
		case ReplaceWith:
			if waitforCMD{
                 waitforVariable=true
                 myTasks.Tasks=append(myTasks.Tasks,Task{
                 	Order:counter,
                 	TheTask:ReplaceWith,
				 })
                 str=""
			}

		case Add:
		case Variable:
			myTasks.Value=append(myTasks.Value,k[orcaCounter])
			orcaCounter++


		case " ":
			str=""
		case "0","1","2","3","4","5","6","7","8","9":

			i,err:=strconv.Atoi(str)
			if err!=nil{
				return OLTask{},err
			}
			if fieldassigned==false{
				myTasks.TheFieldNum=i
				fieldassigned=true
			}

			myTasks.Tasks=append(myTasks.Tasks,Task{
				Order:counter,
				TheTask:Dot,
			})
			myTasks.Value=append(myTasks.Value,"")
			str=""
			waitforDot=false
			waitforCMD=true

		default:




		}
	}



	return myTasks,nil
}


func SetGlobalHook(orcaLanguage string , s ...interface{}){



}

func Start(){

 //start client and server


	tmpAddr:=os.TempDir()
	pid:=os.Getpid()

	if _,err:=os.Stat(tmpAddr+"/Orca"+strconv.Itoa(pid)+".sock");err==nil{
		//file exist
		os.Remove(tmpAddr+"/LeaderOrca.sock")
	}

	l,err:= net.Listen("unix",tmpAddr+"/Orca"+strconv.Itoa(pid)+".sock")
	Check(err)

	go startServerForNormalOrca(l)

	if _,err:=os.Stat(tmpAddr+"/Orca2"+strconv.Itoa(pid)+".sock");err==nil{
		//file exist
		os.Remove(tmpAddr+"/LeaderOrca.sock")
	}

	l2,err:= net.Listen("unix",tmpAddr+"/Orca2"+strconv.Itoa(pid)+".sock")
	Check(err)

	go startServerForNormalOrcatoMessaging(l2)

	//now orca send pid number to leader orca

	msg:=HelloOrca{
		MyPid:pid,
		Flag:0,//0 means register
	}

	go startClientandSendMessage(msg)

}

func startClientandSendMessage(orcaSays HelloOrca){

	tmpAddr:=os.TempDir()
	var network bytes.Buffer
	c,err:=net.Dial("unix",tmpAddr+"/LeaderOrca2.sock")
	Check(err)
	gob.Register(HelloOrca{})
	enc:=gob.NewEncoder(c)
	enc.Encode(orcaSays)

	_,err=c.Write(network.Bytes())
	Check(err)
}

func startClient(pid int,task OLTask){

	tmpAddr:=os.TempDir()
	var network bytes.Buffer
	c,err:=net.Dial("unix",tmpAddr+"/Orca"+strconv.Itoa(pid)+".sock")
	Check(err)
	gob.Register(OLTask{})
	enc:=gob.NewEncoder(c)
	enc.Encode(task)

	_,err=c.Write(network.Bytes())
	Check(err)




}
type HelloOrca struct {
	MyPid int
	Flag int
}

func startServerForNormalOrcatoMessaging(listener net.Listener){
	for {
		fd , err:=listener.Accept()
		Check(err)
		gob.Register(HelloOrca{})
		dec := gob.NewDecoder(fd)
		var v HelloOrca
		dec.Decode(&v)
		fmt.Println(v)

	}
}

func startServerForNormalOrca(listener net.Listener){
	for {
		fd , err:=listener.Accept()
		Check(err)
		gob.Register(OLTask{})
		dec := gob.NewDecoder(fd)
		var v OLTask
		dec.Decode(&v)
		fmt.Println(v)

	}
}
var globalHook OLTask
var pidOfOrcas []int
var mux sync.Mutex
func startServerForMessage(listener net.Listener){
	for {
		fd , err:=listener.Accept()
		Check(err)
		gob.Register(HelloOrca{})
		dec := gob.NewDecoder(fd)
		var v HelloOrca
		dec.Decode(&v)

		if v.Flag==0{
			//register operation
			mux.Lock()
			pidOfOrcas=append(pidOfOrcas,v.MyPid)
			mux.Unlock()
			startClient(v.MyPid,globalHook)

		}

		fmt.Println(v)

	}
}

func startServer(listener net.Listener){
	for {
		fd , err:=listener.Accept()
		Check(err)
		gob.Register(OLTask{})
		dec := gob.NewDecoder(fd)
		var v OLTask
		dec.Decode(&v)
		fmt.Println(v)

	}
}

func StartAsLeader(){
  //start client and server

	/*  var s Scanner

		   sonuc,err:= s.Read("x.9 ReplaceWith {Orca}","merhaba")
		   Check(err)
		   mux.Lock()
		   globalHook=sonuc
		   mux.Unlock()


	tmpAddr:=os.TempDir()

	if _,err:=os.Stat(tmpAddr+"/LeaderOrca.sock");err==nil{
		//file exist
		os.Remove(tmpAddr+"/LeaderOrca.sock")
	}

	l,err:= net.Listen("unix",tmpAddr+"/LeaderOrca.sock")
	Check(err)

	go startServer(l)

	if _,err:=os.Stat(tmpAddr+"/LeaderOrca2.sock");err==nil{
		//file exist
		os.Remove(tmpAddr+"/LeaderOrca2.sock")
	}

	l2,err:=net.Listen("unix",tmpAddr+"/LeaderOrca2.sock")
	Check(err)

	go startServerForMessage(l2)



}


 */



















