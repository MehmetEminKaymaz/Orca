package Orca

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"sort"
)



type MClient struct {
   Db *mongo.Client
   DbOps MongoDBOptions
   LocalH []LocalHook
}

type MongoCollection struct {
	MCollection *mongo.Collection
	MyDb *mongo.Client
	ListId []interface{}
	List []interface{}
	LocalH []LocalHook
}

type MongoDBOptions struct {
	ApplyURI string
	DatabaseName string
}

func(m MongoDBOptions) Options() []string{
	var ops []string
	ops = append(ops,m.ApplyURI)
	ops = append(ops,m.DatabaseName)

	return ops
}

func SetMongoDBOptions(applyURI , DatabaseName string) IOptions{
	return MongoDBOptions{
		ApplyURI:applyURI,
		DatabaseName:DatabaseName,
	}
}

func getDatabaseMongo(applyURI , DatabaseName string) *MClient {
	client,err:=mongo.NewClient(options.Client().ApplyURI(applyURI))
	Check(err)
    //client conenct to server
    err = client.Connect(context.Background())
    Check(err)
    //
	//check the connection
	err = client.Ping(context.TODO(),nil)
	Check(err)
	//
	return &MClient{
		Db:client,
		DbOps:MongoDBOptions{
			ApplyURI:applyURI,
			DatabaseName:DatabaseName,
		},
		LocalH:[]LocalHook{},

	}
}
//will be implement immediately

func(m *MClient) AddLocalHooks(hks ...LocalHook){

	var ids []string
	for i:=0;i< len(hks);i++{
		ids = append(ids,hks[i].getID())
	}
	m.DeleteLocalHooks(ids...)
	m.LocalH=append(m.LocalH,hks...)


}
func(m *MClient) AddLocalHook(hks LocalHook){
	m.DeleteLocalHook(hks.getID())
	m.LocalH=append(m.LocalH,hks)
}
func(m *MClient) DeleteLocalHook(hks string){
	for i:=0;i< len(m.LocalH);i++{
		if m.LocalH[i].getID()==hks{
			m.LocalH[i]=m.LocalH[len(m.LocalH)-1]
			m.LocalH=m.LocalH[:len(m.LocalH)-1]
			break
		}
	}
}
func(m *MClient) DeleteLocalHooks(hks ...string){

	m.LocalH=reorder(m.LocalH,hks)
}

//
func (m *MClient) GetCollection(x interface{},collectionName string ) ICollection{


	collection:=m.Db.Database(m.DbOps.DatabaseName).Collection(collectionName)



	if countN,err:=collection.CountDocuments(context.Background(),bson.D{});countN!=0 {
		var list []interface{}
		var listID []interface{}
		Check(err)

		newSlice:=reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(x)),int(countN),int(countN))

		theSlice:=reflect.New(newSlice.Type())
		theSlice.Elem().Set(newSlice)

		//auto select !

		cursor, err := collection.Find(context.TODO(), bson.D{})

		err=cursor.All(context.TODO(),theSlice.Interface())

		//fmt.Println(theSlice)

		Check(err)

		for i:=0;i<reflect.Indirect(theSlice).Len();i++{
			list=append(list,reflect.Indirect(theSlice).Index(i).Interface())
			s,err:=bson.Marshal(reflect.Indirect(theSlice).Index(i).Interface())
			Check(err)
			var m bson.M
			err=collection.FindOne(context.Background(),s).Decode(&m)
			Check(err)
			listID=append(listID,m["_id"])
		}


		return &MongoCollection{
			MCollection:collection,
			List:list,
			ListId:listID,
			MyDb:m.Db,
			LocalH:m.LocalH,
		}
	}



	list:=make([]interface{},0)
	listID:=make([]interface{},0)

	return &MongoCollection{
		MCollection:collection,
		List:list,
		ListId:listID,
		MyDb:m.Db,
		LocalH:m.LocalH,
	}
}

func (mc *MongoCollection) Add(x interface{})  {

	if len(mc.LocalH)>0 {
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

	err:=mc.MyDb.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err:=sessionContext.StartTransaction()
		if err!=nil{
			Check(err)
			return err
		}

		insertResult,err := mc.MCollection.InsertOne(context.TODO(),x)
		if err!=nil{
			sessionContext.AbortTransaction(sessionContext)
			return err
		}else{
			sessionContext.CommitTransaction(sessionContext)
			_=insertResult
			mc.ListId=append(mc.ListId,insertResult.InsertedID)
			mc.List=append(mc.List,x)
		}
		return nil
	})


    Check(err)

	if len(mc.LocalH)>0 {
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


func (mc *MongoCollection) AddRange (x interface{}){

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



	err:=mc.MyDb.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err:=sessionContext.StartTransaction()
		if err!=nil{
			return err
		}

		mslice :=reflect.ValueOf(x)

		var addThis []interface{}


		for i:=0;i<mslice.Len();i++{
			addThis=append(addThis,mslice.Index(i).Interface())
		}

		insertManyResult,err:=mc.MCollection.InsertMany(context.TODO(),addThis)
		if err!=nil{
			sessionContext.AbortTransaction(sessionContext)
			return err
		}else{
			sessionContext.CommitTransaction(sessionContext)
			_=insertManyResult

			mc.ListId=append(mc.ListId,insertManyResult.InsertedIDs...)
			mc.List=append(mc.List,addThis...)
		}
		return nil
	})




	Check(err)

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
func (mc *MongoCollection) Update(x interface{},y interface{}){


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
			x = funk(x)
		}
		//local hook
	}


	err:=mc.MyDb.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err:=sessionContext.StartTransaction()
		if err!=nil{
			return err
		}

		oldbayt,err:=bson.Marshal(x)
		Check(err)
		newbayt,err:=bson.Marshal(y)
		Check(err)

		updateResult,err:=mc.MCollection.ReplaceOne(context.TODO(),oldbayt,newbayt)

		if err!=nil{
			sessionContext.AbortTransaction(sessionContext)
			return err
		}else{
			sessionContext.CommitTransaction(sessionContext)
			_=updateResult

			//update does not change objectId
			for i:=0;i< len(mc.List);i++{
				if mc.List[i]==x{
					mc.List[i]=y
					break
				}
			}

		}
		return nil
	})






	Check(err)


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
			funk(y)
		}
		//local hook
	}


}
func (mc *MongoCollection) Delete(x interface{}){

	if len(mc.LocalH)>0 {
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



	err:=mc.MyDb.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err:=sessionContext.StartTransaction()
		if err!=nil{
			return err
		}

		bayt , err :=bson.Marshal(x)
		Check(err)

		deleteResult ,err :=mc.MCollection.DeleteOne(context.TODO(),bayt)

		if err!=nil{
			sessionContext.AbortTransaction(sessionContext)
			return err
		}else{
			sessionContext.CommitTransaction(sessionContext)
			_=deleteResult

			for i:=0;i< len(mc.List);i++{
				if mc.List[i]==x{
					mc.List[len(mc.List)-1],mc.List[i]=mc.List[i],mc.List[len(mc.List)-1]
					mc.List=mc.List[:len(mc.List)-1]
					mc.ListId[len(mc.ListId)-1],mc.ListId[i]=mc.ListId[i],mc.ListId[len(mc.ListId)-1]
					mc.ListId=mc.ListId[:len(mc.ListId)-1]
					break
				}
			}
		}
		return nil

	})


	Check(err)

	if len(mc.LocalH)>0 {
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
func (mc *MongoCollection) Clear(){

	err:=mc.MyDb.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err:=sessionContext.StartTransaction()
		if err!=nil{
			return err
		}

		deleteResult , err := mc.MCollection.DeleteMany(context.TODO(),bson.D{{}})

		if err!=nil{
			sessionContext.AbortTransaction(sessionContext)
			return err
		}else{
			sessionContext.CommitTransaction(sessionContext)
			_=deleteResult

			mc.ListId=mc.ListId[:0]
			mc.List=mc.List[:0]
		}
		return nil
	})



     Check(err)


}
func (mc *MongoCollection) Foreach(x interface{}){ //must be support



}
func (mc *MongoCollection) GetLogs(){//must be support

}

func (mc *MongoCollection) ToSlice() []interface{}{

	slice :=make([]interface{},0,1)

	slice = append(slice,mc.List...)

	return slice
}




