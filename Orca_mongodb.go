package Orca

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
)



type MClient struct {
   Db *mongo.Client
   DbOps MongoDBOptions
}

type MongoCollection struct {
	MCollection *mongo.Collection
	MyDb *mongo.Client
	ListId []interface{}
	List []interface{}
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

	}
}

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
		}
	}



	list:=make([]interface{},0)
	listID:=make([]interface{},0)

	return &MongoCollection{
		MCollection:collection,
		List:list,
		ListId:listID,
		MyDb:m.Db,
	}
}

func (mc *MongoCollection) Add(x interface{})  {

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


}


func (mc *MongoCollection) AddRange (x interface{}){

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


}
func (mc *MongoCollection) Update(x interface{},y interface{}){


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


}
func (mc *MongoCollection) Delete(x interface{}){


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




