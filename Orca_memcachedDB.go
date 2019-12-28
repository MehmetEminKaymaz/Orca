package Orca

import (
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"reflect"
)

type NewMemCacheVariable struct{
	Key string
	Value interface{}
	ExpirationTime int
}

type MemCachedDB struct {
	Db *memcache.Client
	DbOps MemCachedOptions

}

type MemCachedCollection struct{
	MyDb *memcache.Client
	Separator string
	RelationName string
    KeyList []string
	ValueList []interface{}
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
	}

}

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
	}

}

func(mc *MemCachedCollection) Add(x interface{}){
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
}

func(mc *MemCachedCollection) AddRange(x interface{}){

	switch x.(type) {
	case []NewMemCacheVariable,[]NewRedisVariable:
		val:=reflect.ValueOf(x)
		for i:=0;i<val.Len();i++{
			mc.Add(val.Index(i))
		}
	default:

	}

}
func(mc *MemCachedCollection) Update(old interface{},New interface{}){

	mc.Delete(old)
	mc.Add(New)

}
func(mc *MemCachedCollection) Delete(x interface{}){

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


