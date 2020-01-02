package Orca

import (
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"reflect"
	"sort"
)

type NewMemCacheVariable struct{
	Key string
	Value interface{}
	ExpirationTime int
}

type MemCachedDB struct {
	Db *memcache.Client
	DbOps MemCachedOptions
	LocalH []LocalHook
    lock bool
}

type MemCachedCollection struct{
	MyDb *memcache.Client
	Separator string
	RelationName string
    KeyList []string
	ValueList []interface{}
	LocalH []LocalHook
	lock bool
}

type MemCachedOptions struct {
	Servers []string
}

func(m MemCachedOptions) Options() []string{
	return m.Servers
}

func SetMemCachedDBOptions(Servers ...string) IOptions{

	return MemCachedOptions{
		Servers:Servers,
	}
}

func getDatabaseMemCached(Servers []string) *MemCachedDB{
	client := memcache.New(Servers...)

	return &MemCachedDB{
		Db:client,
		DbOps:MemCachedOptions{
			Servers:Servers,
		},
		LocalH:[]LocalHook{},
		lock:false,
	}

}
//will be implement immediately

func(m *MemCachedDB) AddLocalHooks(hks ...LocalHook){

	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	m.DeleteLocalHooks(ids...)
	m.LocalH=append(m.LocalH,hks...)

}
func(m *MemCachedDB) AddLocalHook(hks LocalHook){

	m.DeleteLocalHook(hks.getID())
	m.LocalH=append(m.LocalH,hks)
}
func(m *MemCachedDB) DeleteLocalHook(hks string){
	for i:=0;i< len(m.LocalH);i++{
		if m.LocalH[i].getID()==hks{
			m.LocalH[i]=m.LocalH[len(m.LocalH)-1]
			m.LocalH=m.LocalH[:len(m.LocalH)-1]
			break
		}
	}
}
func(m *MemCachedDB) DeleteLocalHooks(hks ...string){
	m.LocalH=reorder(m.LocalH,hks)
}

//


func(m *MemCachedDB) GetCollection(x interface{},relationName string) ICollection{

	//auto select!



	//



	if x.(string)==""{
		x=":"
	}

	_=relationName

	return &MemCachedCollection{
		MyDb:m.Db,
		Separator:x.(string),
		RelationName:relationName,
		KeyList:make([]string,0),
		ValueList:make([]interface{},0),
		LocalH:m.LocalH,
		lock:false,
	}

}

func(mc *MemCachedCollection) Add(x interface{}){

	if mc.lock==false {

		if len(mc.LocalH) > 0 {
			//local hook
			var beforeAddLocalHooks []LocalHook
			for i := 0; i < len(mc.LocalH); i++ {
				if _, n := mc.LocalH[i].getSign(); n == BeforeAdd {
					beforeAddLocalHooks = append(beforeAddLocalHooks, mc.LocalH[i])
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

	switch x.(type) {
	case NewMemCacheVariable,NewRedisVariable:
		val:=reflect.ValueOf(x)
		//look if exist update the key!
		//
		//else (so if not exist) add a new key

		//to database
		encoded,err:=json.Marshal(val.Field(1).Interface())
		Check(err)

		err=mc.MyDb.Add(&memcache.Item{
			Key:mc.RelationName+mc.Separator+val.Field(0).String(),
			Value:encoded,
			Expiration:int32(val.Field(2).Int()),
		})

		//to lists
		if err==nil{
			mc.KeyList=append(mc.KeyList,val.Field(0).String())
			mc.ValueList=append(mc.ValueList,val.Field(1).Interface())
		}

		Check(err)


		//
	default:


	}

	if mc.lock==false {
		if len(mc.LocalH) > 0 {
			//local hook
			var afterAddLocalHooks []LocalHook
			for i := 0; i < len(mc.LocalH); i++ {
				if _, n := mc.LocalH[i].getSign(); n == AfterAdd {
					afterAddLocalHooks = append(afterAddLocalHooks, mc.LocalH[i])
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

func(mc *MemCachedCollection) AddRange(x interface{}){


	if len(mc.LocalH)>0 {
		//local hook
		var beforeAddRangeLocalHooks []LocalHook
		for i := 0; i < len(mc.LocalH); i++ {
			if _, n := mc.LocalH[i].getSign(); n == BeforeAddRange {
				beforeAddRangeLocalHooks = append(beforeAddRangeLocalHooks, mc.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeAddRangeLocalHooks))

		for i := 0; i < len(beforeAddRangeLocalHooks); i++ {
			funk := beforeAddRangeLocalHooks[i].getHookFunc()
			x = funk(x)
		}
		//local hook
	}

	switch x.(type) {
	case []NewMemCacheVariable,[]NewRedisVariable:
		val:=reflect.ValueOf(x)
		for i:=0;i<val.Len();i++{
			mc.Add(val.Index(i))
		}
	default:

	}


	if len(mc.LocalH)>0 {
		//local hook
		var afterAddRangeLocalHooks []LocalHook
		for i := 0; i < len(mc.LocalH); i++ {
			if _, n := mc.LocalH[i].getSign(); n == AfterAddRange {
				afterAddRangeLocalHooks = append(afterAddRangeLocalHooks, mc.LocalH[i])
			}
		}

		sort.Sort(byPriority(afterAddRangeLocalHooks))

		for i := 0; i < len(afterAddRangeLocalHooks); i++ {
			funk := afterAddRangeLocalHooks[i].getHookFunc()
			funk(x)
		}
		//local hook
	}

}
func(mc *MemCachedCollection) Update(old interface{},New interface{}){

    mc.lock=true
	if len(mc.LocalH)>0 {
		//local hook
		var beforeUpdateLocalHooks []LocalHook
		for i := 0; i < len(mc.LocalH); i++ {
			if _, n := mc.LocalH[i].getSign(); n == BeforeUpdate {
				beforeUpdateLocalHooks = append(beforeUpdateLocalHooks, mc.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeUpdateLocalHooks))

		for i := 0; i < len(beforeUpdateLocalHooks); i++ {
			funk := beforeUpdateLocalHooks[i].getHookFunc()
			old = funk(old)
		}
		//local hook

	}


	mc.Delete(old)
	mc.Add(New)

	if len(mc.LocalH)>0 {
		//local hook
		var afterUpdateLocalHooks []LocalHook
		for i := 0; i < len(mc.LocalH); i++ {
			if _, n := mc.LocalH[i].getSign(); n == AfterUpdate {
				afterUpdateLocalHooks = append(afterUpdateLocalHooks, mc.LocalH[i])
			}
		}

		sort.Sort(byPriority(afterUpdateLocalHooks))

		for i := 0; i < len(afterUpdateLocalHooks); i++ {
			funk := afterUpdateLocalHooks[i].getHookFunc()
			funk(New)
		}
		//local hook
	}

    mc.lock=false

}
func(mc *MemCachedCollection) Delete(x interface{}){

	if mc.lock==false {
		if len(mc.LocalH) > 0 {
			//local hook
			var beforeDeleteLocalHooks []LocalHook
			for i := 0; i < len(mc.LocalH); i++ {
				if _, n := mc.LocalH[i].getSign(); n == BeforeDelete {
					beforeDeleteLocalHooks = append(beforeDeleteLocalHooks, mc.LocalH[i])
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

	switch x.(type) {
	case NewMemCacheVariable,NewRedisVariable:

		val:=reflect.ValueOf(x)

		err:=mc.MyDb.Delete(mc.RelationName+mc.Separator+val.Field(0).String())
		if err==nil{
			str:=val.Field(0).String()
			for i:=0;i< len(mc.KeyList);i++{
				if mc.KeyList[i]==str{
					mc.KeyList[len(mc.KeyList)-1],mc.KeyList[i]=mc.KeyList[i],mc.KeyList[len(mc.KeyList)-1]
					mc.KeyList=mc.KeyList[:len(mc.KeyList)-1]
					mc.ValueList[len(mc.ValueList)-1],mc.ValueList[i]=mc.ValueList[i],mc.ValueList[len(mc.ValueList)-1]
					mc.ValueList=mc.ValueList[:len(mc.ValueList)-1]
					break
				}
			}
		}
		Check(err)

	default:

	}

	if mc.lock==false {
		if len(mc.LocalH) > 0 {
			//local hook
			var afterDeleteLocalHooks []LocalHook
			for i := 0; i < len(mc.LocalH); i++ {
				if _, n := mc.LocalH[i].getSign(); n == AfterDelete {
					afterDeleteLocalHooks = append(afterDeleteLocalHooks, mc.LocalH[i])
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
func(mc *MemCachedCollection) Clear(){

	for i:=0;i< len(mc.KeyList);i++{
		err:=mc.MyDb.Delete(mc.RelationName+mc.Separator+mc.KeyList[i])
		if err==nil{
			mc.KeyList[len(mc.KeyList)-1],mc.KeyList[i]=mc.KeyList[i],mc.KeyList[len(mc.KeyList)-1]
			mc.KeyList=mc.KeyList[:len(mc.KeyList)-1]
			mc.ValueList[len(mc.ValueList)-1],mc.ValueList[i]=mc.ValueList[i],mc.ValueList[len(mc.ValueList)-1]
			mc.ValueList=mc.ValueList[:len(mc.ValueList)-1]
		}
		Check(err)

	}

}
func(mc *MemCachedCollection) Foreach(x interface{}){

}
func(mc *MemCachedCollection) GetLogs(){

}
func(mc *MemCachedCollection) ToSlice() []interface{}{
	return mc.ValueList
}


func MemCacheDBClient(c ICollection) *MemCachedCollection{

    return c.(*MemCachedCollection)

}


func MemCacheDBGet( mc *MemCachedCollection , key string) interface{}{


	var v interface{}
	item , err:=mc.MyDb.Get(key)
	Check(err)
	err=json.Unmarshal(item.Value,&v)

	return v

}


