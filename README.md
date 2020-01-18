![Image of Orca](https://github.com/MehmetEminKaymaz/Orca/blob/master/Orca.png)

# Orca
Simple and Powerful ORM for  Go Language (MongoDB,SQLite,RedisDB and Memcache)


### Install
```
 go get github.com/MehmetEminKaymaz/Orca
 
```

Define your structs

```Go

type Person struct {
	Name string
	Age int
        MyHead Head
	FriendsName []string
}

type Head struct {

	Value Brain

}

type Brain struct {

	IQ int

}


```

then use Orca to perform CRUD operations

# Database Operations

Firstly create some variables.

```Go
//create a person
	someone:=Person{
		Age:99,
		Name:"Gandalf",
		FriendsName: []string{"Frodo","Sam","Galadriel","Radagast"},
		MyHead:Head{
			Value:Brain{
				IQ:1000,//:)
			},
		},
	}
	
	newPerson:=someone
	newPerson.Name="Saruman"
	
```

## SQLite

```Go

SQLiteDB,err:=Orca.Use(Orca.SQLite,Orca.SetSQLiteOptions("sqlite3","./MyDb.db")) //get database
_=err
myPersonList:=SQLiteDB.GetCollection(Person{},"Person") 
//if there is no table in the database to store a person, it creates it
//if there is a table, it goes to the table and returns a collection of person

myPersonList.Add(someone)//it inserts person to database

myPersonList.Update(someone,newPerson) //updated! 

myPersonList.Delete(newPerson)//newPerson deleted from database.

myPersonList.Clear()//it delete all tuples in table(person table).

myPersonList.ToSlice() //it return a slice of person

```

## MongoDB

```Go
MyMongoDB,err:=Orca.Use(Orca.MongoDB,Orca.SetMongoDBOptions("mongodb://127.0.0.1","myMongo")) //get database
_=err

myPersonList:=MyMongoDB.GetCollection(Person{},"Person")
//if there is no collection in the database to store a person, it creates it
//if there is a collection, it goes to the collection and returns a collection of person

myPersonList.Add(someone)//it inserts person to database

myPersonList.AddRange(SliceOfPerson)//it inserts multiple person in one transaction

myPersonList.Update(someone,newPerson) //updated!

myPersonList.Delete(newPerson)//newPerson deleted from database.

myPersonList.Clear()//it delete all tuples in table(person table).

myPersonList.ToSlice() //it return a slice of person

```

## RedisDB

```Go
MyRedisDB,err:=Orca.Use(Orca.Redis,Orca.SetRedisDBOptions("localhost:6379","","0")) //get database
_=err

myPersonList:=MyRedisDB.GetCollection(":","Person") //first parameter is seperator , second is relation

myPersonList.Add(Orca.NewRedisVariable{
Key :"150316",
Value : "Gandalf",
ExpirationTime:0,
}) //it insert to database as => Key =  Person:150316 and Value = Gandalf 
//so , result of the 'get "Person:150316"' code is  "Gandalf".

//if you do not want to use relation and seperator set empty all , example : 
myPersonList:=MyRedisDB.GetCollection("","")

myPersonList.AddRange([]Orca.NewRedisVariable{...}) //add multiple key and values to database

myPersonList.Delete(...) //it delete from database.
myPersonList.Clear() //it deletes only person if you use relation , otherwise it delete all key-value pairs
myPerson.Update(oldValue,newValue)//updated!


```
## Memcache

Orca also support Add,AddRange,Clear,Delete,Update method for Memcache DB.The only difference between Redis and memcache usage is Orca.NewMemCacheVariable declaration for key-value pairs.But you can use Orca.NewMemCacheVariable in Redis operations or Orca.NewRedisVariable in Memcache operations.

Example : 
```Go

MemCachedDB,err:=Orca.Use(Orca.MemCached,Orca.SetMemCachedDBOptions("localhost:11211"))
_=err
myPersonList :=MemCachedDB.GetCollection(":","Person") //same features (seperator and relation)

myPersonList.Add(Orca.NewRedisVariable{..})//correct!
myPersonList.Add(Orca.NewMemCacheVariable{..})//correct!

//others same...(AddRange,Clear,Update,Delete,ToSlice etc.)

```

# LocalHook Usage in ORCA

## Orca offers the programmer the following localhooks : 

 * BeforeAdd
 * AfterAdd
 * BeforeAddRange
 * AfterAddRange
 * BeforeUpdate
 * AfterUpdate
 * BeforeDelete
 * AfterDelete
 
## What is LocalHook ? 

 If you need run a method/function inside crud operation functions(Add,AddRange,Delete etc.) , you must use LocalHooks.
 
 Orca.NewLocalHook :   
  First parameter is LocalHook id (like key) , if you need to find your hook you need the key.  
  Second parameter is Priority you can add many BeforeAdd hook , Orca execute all hooks in order.  
  0 is Higher priority.  
  Third is Hook location (beforeAdd,afterAdd, etc.)  
  The last is function.  


## LocalHook Example with SQLite 

```Go
SQLiteDB,err:=Orca.Use(Orca.SQLite,Orca.SetSQLiteOptions("sqlite3","./MyDb.db")) //get database
_=err

SQLiteDB.AddLocalHook(Orca.NewLocalHook(
		"myHook",0,Orca.BeforeAdd, func(i interface{}) interface{} {
			return i
		}))//add BeforeAdd localhook.
//The function run before all add operations and it takes the value to be added parameter as empty interface.
//you can add the other hooks using the same way.


SQLiteDB.AddLocalHooks(...)//add multiple localhooks at same time.
SQliteDB.DeleteHook(string)//delete hook using id
SQliteDB.DeleteHooks(...string)//delete hooks

 
```

LocalHook usage is same for the other databases (RedisDB,MongoDB,Memcache...)

# Generic Tools in ORCA
Orca offers 4 type of generic tool to manipulate results without change database.
These Tools are : 
 * LazyList : LazyList is lazy :) it does not perform until call Do function.It stores functions to execute after called do function.
 * LinkedList :  A list implemented by using go/list package.
 * ImmutableList : ImmutableList is easy to make thread safe operations.Methods access the copy of the source.
 * MutableList :
 MutableList methods access the source using pointers.
  
  

Example : 
```Go
        ImmutableList.NewImmutableList(myPersonList.ToSlice())
	MutableList.NewMutableList(myPersonList.ToSlice())
	LazyList.NewLazyList(myPersonList.ToSlice())
	LinkedList.NewLinkedList(myPersonList.ToSlice())
	
```

```Go
       mylist:=MutableList.NewMutableList(myPersonList.ToSlice())
	    mylist.Where(func(x interface{}) bool {
		return x.(Person).Age<30
	    })//it returns slice of person where age smaller than 30
	    
	    //and more (Contains,Distict,ElementAt,Exist,Foreach,RemoveAt,Skip,Take etc..) 
```

Finally print the result : 

```
       fmt.println(mylist.ToSlice())
       
```






