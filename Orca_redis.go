package Orca

import (
	"github.com/go-redis/redis"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type NewRedisVariable struct {

	Key interface{}
	Value interface{}
	ExpirationTime interface{}

}

type RedisClient struct {
	Db *redis.Client
    DbOps RedisOptions
	LocalH []LocalHook
	lock bool
}

type RedisCollection struct {
	MyDb *redis.Client
	Separator string
	TableName string
	KeysList []interface{}
	ValueList []interface{}
	LocalH []LocalHook
	lock bool

}

type RedisOptions struct {
	Addr string
	Password string
	DB string
}


func (r RedisOptions) Options() []string{
	var ops []string

	ops = append(ops,r.Addr)
	ops=append(ops,r.Password)
	ops = append(ops,r.DB)

	return ops

}
//will be implement immediately
func(r *RedisClient) AddLocalHooks(hks ...LocalHook){

	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	r.DeleteLocalHooks(ids...)
	r.LocalH=append(r.LocalH,hks...)


}
func(r *RedisClient) AddLocalHook(hks LocalHook){
	r.DeleteLocalHook(hks.getID())
	r.LocalH=append(r.LocalH,hks)
}
func(r *RedisClient) DeleteLocalHook(hks string){

	for i:=0;i< len(r.LocalH);i++{
		if r.LocalH[i].getID()==hks{
			r.LocalH[i]=r.LocalH[len(r.LocalH)-1]
			r.LocalH=r.LocalH[:len(r.LocalH)-1]
			break
		}
	}

}
func(r *RedisClient) DeleteLocalHooks(hks ...string){

	r.LocalH=reorder(r.LocalH,hks)

}


//

func (r *RedisClient) GetCollection(x interface{},tableName string) ICollection{

	//auto select !

	relatedKeys,err:=r.Db.Keys("*"+tableName+"*").Result()
	var fixedRelatedKeys []string
	for i:=0;i< len(relatedKeys);i++{

		fixedRelatedKeys=append(fixedRelatedKeys,strings.Replace(relatedKeys[i],tableName+x.(string),"",-1))
	}

	if len(relatedKeys)==0{
		return &RedisCollection{
			MyDb:r.Db,
			Separator:x.(string),
			TableName:tableName,
			KeysList:make([]interface{},0),
			ValueList:make([]interface{},0),
			LocalH:r.LocalH,
			lock:false,
		}
	}
	Check(err)
	relatedValues,err:=r.Db.MGet(relatedKeys...).Result()
	Check(err)
	var Keys []interface{}
	for i:=0;i< len(fixedRelatedKeys);i++{
		Keys=append(Keys,fixedRelatedKeys[i])
	}
	//auto select end

	if x.(string)==""{
		x=":"
	}

	_=tableName

	return &RedisCollection{
		MyDb:r.Db,
		Separator:x.(string),
		TableName:tableName,
		KeysList:Keys,
		ValueList:relatedValues,
		LocalH:r.LocalH,
		lock:false,

	}
}

func SetRedisDBOptions(Addr,Password,DB string) IOptions{
	return RedisOptions{
		DB:DB,
		Password:Password,
		Addr:Addr,
	}
}

func getDatabaseRedis(Addr,Password,DB string) *RedisClient{
	db,err:=strconv.Atoi(DB)
	Check(err)

	client:=redis.NewClient(&redis.Options{
		Addr:Addr,
		Password:Password,
		DB:db,
	})

	return &RedisClient{
		Db:client,
		DbOps:RedisOptions{
			DB:DB,
			Password:Password,
			Addr:Addr,
		},
		LocalH:[]LocalHook{},
		lock:false,
	}


}

func (r *RedisCollection) Add(x interface{}){

	if r.lock==false {

		if len(r.LocalH) > 0 {

			var beforeAddLocalHooks []LocalHook
			for i := 0; i < len(r.LocalH); i++ {
				if _, n := r.LocalH[i].getSign(); n == BeforeAdd {
					beforeAddLocalHooks = append(beforeAddLocalHooks, r.LocalH[i])
				}
			}

			sort.Sort(byPriority(beforeAddLocalHooks))

			for i := 0; i < len(beforeAddLocalHooks); i++ {
				funk := beforeAddLocalHooks[i].getHookFunc()
				x = funk(x)
			}
		}

	}

	switch x.(type) {
	case NewRedisVariable,NewMemCacheVariable:

		if t:=isThereAKey(r.KeysList,reflect.ValueOf(x).Field(0).Interface()); t{
			//there is a key
			for i:=0;i< len(r.KeysList);i++{
				if r.KeysList[i]==reflect.ValueOf(x).Field(0).Interface(){ //if this is the key we are looking for
					r.ValueList[i]=reflect.ValueOf(x).Field(1).Interface() //set the new value to valueList
					break //we do not need to iterate anymore
				}
			}
		}else{
			//there is not a key

			r.KeysList=append(r.KeysList,reflect.ValueOf(x).Field(0).Interface())
			r.ValueList=append(r.ValueList,reflect.ValueOf(x).Field(1).Interface())

		}


		err:=r.MyDb.Watch(func(tx *redis.Tx) error {
			n,err:=tx.Get(r.TableName+r.Separator+reflect.ValueOf(x).Field(0).Interface().(string)).Result()
			if err!=nil&&err!=redis.Nil{
				return err
			}
			_,err=tx.Pipelined(func(pipeliner redis.Pipeliner) error {
				pipeliner.Set(r.TableName+r.Separator+reflect.ValueOf(x).Field(0).Interface().(string),reflect.ValueOf(x).Field(1).Interface(),time.Duration(reflect.ValueOf(x).Field(2).Interface().(int)))
				return nil
			})
			_ = n
			return err

		},r.TableName+r.Separator+reflect.ValueOf(x).Field(0).Interface().(string))
		Check(err)
	default:


	}

	if r.lock==false {

		if len(r.LocalH) > 0 {
			//local hook
			var afterAddLocalHooks []LocalHook
			for i := 0; i < len(r.LocalH); i++ {
				if _, n := r.LocalH[i].getSign(); n == AfterAdd {
					afterAddLocalHooks = append(afterAddLocalHooks, r.LocalH[i])
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

func isThereAKey(slice interface{},key interface{}) bool{

	for i:=0;i<reflect.ValueOf(slice).Len();i++{

		if reflect.ValueOf(slice).Index(i).Interface()==key{

			return true

		}

	}
	return false
}

func (r *RedisCollection) AddRange(x interface{}){

	if len(r.LocalH)>0 {
		//local hook
		var beforeAddRangeLocalHooks []LocalHook
		for i := 0; i < len(r.LocalH); i++ {
			if _, n := r.LocalH[i].getSign(); n == BeforeAddRange {
				beforeAddRangeLocalHooks = append(beforeAddRangeLocalHooks, r.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeAddRangeLocalHooks))

		for i := 0; i < len(beforeAddRangeLocalHooks); i++ {
			funk := beforeAddRangeLocalHooks[i].getHookFunc()
			x = funk(x)
		}
		//local hook
	}


	switch  x.(type){

	case []NewRedisVariable,[]NewMemCacheVariable:

		ourSlice := reflect.ValueOf(x)

		for k:=0;k<ourSlice.Len();k++{

			if t:=isThereAKey(r.KeysList,reflect.ValueOf(ourSlice.Index(k).Interface()).Field(0).Interface()); t{
				//there is a key
				for i:=0;i< len(r.KeysList);i++{
					if r.KeysList[i]==reflect.ValueOf(ourSlice.Index(k).Interface()).Field(0).Interface(){ //if this is the key we are looking for
						r.ValueList[i]=reflect.ValueOf(ourSlice.Index(k).Interface()).Field(1).Interface() //set the new value to valueList
						break //we do not need to iterate anymore
					}
				}
			}else{
				//there is not a key

				r.KeysList=append(r.KeysList,reflect.ValueOf(ourSlice.Index(k).Interface()).Field(0).Interface())
				r.ValueList=append(r.ValueList,reflect.ValueOf(ourSlice.Index(k).Interface()).Field(1).Interface())

			}

		}





		var Keys []string
		var Values []interface{}
		var ExpirationTimes []time.Duration
		for i:=0;i<ourSlice.Len();i++{
			Keys=append(Keys,r.TableName+r.Separator+ourSlice.Index(i).Field(0).Interface().(string))
			Values = append(Values,ourSlice.Index(i).Field(1).Interface())
			ExpirationTimes = append(ExpirationTimes,time.Duration(ourSlice.Index(i).Field(2).Interface().(int)))
		}

		var pairs []interface{}

		for i:=0;i< len(Keys);i++{
			pairs=append(pairs, Keys[i],Values[i])
		}

		err:=r.MyDb.Watch(func(tx *redis.Tx) error {



			n,err:=tx.MGet(Keys...).Result()
			if err!=nil&&err!=redis.Nil{
				return err
			}
			_,err=tx.Pipelined(func(pipeliner redis.Pipeliner) error {
                pipeliner.MSet(pairs...)
				return nil
			})
			_ = n
			return err

		},Keys...)
		Check(err)

	default:



	}

	if len(r.LocalH)>0 {
		//local hook
		var afterAddRangeLocalHooks []LocalHook
		for i := 0; i < len(r.LocalH); i++ {
			if _, n := r.LocalH[i].getSign(); n == AfterAddRange {
				afterAddRangeLocalHooks = append(afterAddRangeLocalHooks, r.LocalH[i])
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

func (r *RedisCollection) Update(x interface{},y interface{}){

	r.lock=true
	if len(r.LocalH)>0 {
		//local hook
		var beforeUpdateLocalHooks []LocalHook
		for i := 0; i < len(r.LocalH); i++ {
			if _, n := r.LocalH[i].getSign(); n == BeforeUpdate {
				beforeUpdateLocalHooks = append(beforeUpdateLocalHooks, r.LocalH[i])
			}
		}

		sort.Sort(byPriority(beforeUpdateLocalHooks))

		for i := 0; i < len(beforeUpdateLocalHooks); i++ {
			funk := beforeUpdateLocalHooks[i].getHookFunc()
			x = funk(x)
		}
		//local hook
	}

	r.Delete(x)
	r.Add(y)

	if len(r.LocalH)>0 {
		//local hook
		var afterUpdateLocalHooks []LocalHook
		for i := 0; i < len(r.LocalH); i++ {
			if _, n := r.LocalH[i].getSign(); n == AfterUpdate {
				afterUpdateLocalHooks = append(afterUpdateLocalHooks, r.LocalH[i])
			}
		}

		sort.Sort(byPriority(afterUpdateLocalHooks))

		for i := 0; i < len(afterUpdateLocalHooks); i++ {
			funk := afterUpdateLocalHooks[i].getHookFunc()
			funk(y)
		}
		//local hook
	}
	r.lock=false

}

func (r *RedisCollection) Delete(x interface{}){

	if r.lock==false {
		if len(r.LocalH) > 0 {
			//local hook
			var beforeDeleteLocalHooks []LocalHook
			for i := 0; i < len(r.LocalH); i++ {
				if _, n := r.LocalH[i].getSign(); n == BeforeDelete {
					beforeDeleteLocalHooks = append(beforeDeleteLocalHooks, r.LocalH[i])
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

	//r.MyDb.Del(r.TableName+":"+reflect.ValueOf(x).Field(0).Interface().(string))

	if t:=isThereAKey(r.KeysList,reflect.ValueOf(x).Field(0).Interface());t{
		 for i:=0;i< len(r.KeysList);i++{
		 	if r.KeysList[i]==reflect.ValueOf(x).Field(0).Interface(){

		 		r.KeysList[len(r.KeysList)-1],r.KeysList[i]=r.KeysList[i],r.KeysList[len(r.KeysList)-1]
		 		r.KeysList=r.KeysList[:len(r.KeysList)-1]
		 		r.ValueList[len(r.ValueList)-1],r.ValueList[i]=r.ValueList[i],r.ValueList[len(r.ValueList)-1]
		 		r.ValueList=r.ValueList[:len(r.ValueList)-1]
		 		break

			}
		 }
	}else{

		//nothing

	}

	err:=r.MyDb.Watch(func(tx *redis.Tx) error {

		_,err:=tx.Pipelined(func(pipeliner redis.Pipeliner) error {

			pipeliner.Del(r.TableName+r.Separator+reflect.ValueOf(x).Field(0).Interface().(string))

			return nil
		})

		return err

	},reflect.ValueOf(x).Field(0).Interface().(string))
	Check(err)

	if r.lock==false {

		if len(r.LocalH) > 0 {
			//local hook
			var afterDeleteLocalHooks []LocalHook
			for i := 0; i < len(r.LocalH); i++ {
				if _, n := r.LocalH[i].getSign(); n == AfterDelete {
					afterDeleteLocalHooks = append(afterDeleteLocalHooks, r.LocalH[i])
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

func (r *RedisCollection) Clear(){


	var p []string
	for i:=0;i<len(r.KeysList);i++{
		p=append(p,r.TableName+r.Separator+r.KeysList[i].(string))
	}

	err:=r.MyDb.Watch(func(tx *redis.Tx) error {

		_,err:=tx.Pipelined(func(pipeliner redis.Pipeliner) error {

			//pipeliner.Del(r.TableName+r.Separator+reflect.ValueOf(x).Field(0).Interface().(string))
			pipeliner.Del(p...)

			return nil
		})

		return err

	},p...)
	Check(err)

	//r.MyDb.FlushDB()
	/*for i:=0;i< len(r.KeysList);i++{
		r.Delete(r.KeysList[i])
	}*/

	r.KeysList=r.KeysList[:0]
	r.ValueList=r.ValueList[:0]



}

func (r *RedisCollection) Foreach(x interface{}){

}

func (r *RedisCollection) GetLogs(){

}

func(r *RedisCollection) ToSlice() []interface{}{

	result:=make([]interface{},0)
	for i:=0;i< len(r.KeysList);i++{
		result=append(result,NewRedisVariable{
			Key:r.KeysList[i].(string),
			Value:r.ValueList[i],
			ExpirationTime:0,
		})
	}

	return result
}