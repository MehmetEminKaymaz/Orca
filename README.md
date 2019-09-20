![Image of Orca](https://github.com/MehmetEminKaymaz/Orca/blob/master/Orca.png)

# Orca
Simple and Powerful ORM for  Go Language

Firstly define your structs

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

```Go
    db:=Orca.GetDatabase("sqlite3","./MyDb.db") //get database 
    
    myPersonList:=db.GetCollection(Person{},"Person") //it creates if not exist tables to store Person struct ,
    // if exist it return persons as a collection
    
    //now create a person 
    someone:=Person{
		Age:99,
		Name:"Uzun Yaşar",
		FriendsName: []string{"Ahmet","Hasan","Necmi","Batucan"},
		MyHead:Head{
			Value:Brain{
				IQ:1000,//:)
			},
		},
	}
  
  //then insert it to DB
  
  myPersonList.Insert(someone)//inserted!Thats all...
  
  //update example
  newPerson:=someone
	newPerson.Name="Çok uzun yaşar"
  
  myPersonList.Update(someone,newPerson)//Updated!
  
  //if you want to delete it 
  myPersonList.Delete(newPerson)//it deleted from db!It's easy!
  
  //if you want to delete all records from collection use this :
  myPersonList.DeleteALL()//it deletes all records from db and collection!
  
  //if you want get this collection as a slice of person
  myPersonList.ToSlice()
  
  //if you need linq methods like .NET to manipulate the slice(myPersonList.ToSlice()) 
  Orca.From(myPersonList.ToSlice()).Where(func(x interface{}) bool {
		return x.(Person).Age<30
	}).ToSlice()
  
  Orca.From(myPersonList.ToSlice()).Foreach(func(x interface{}) (y interface{}) {
		a:=(x.(Person).Age)+1

		y=Person{Age:a,Name:x.(Person).Name,MyHead:x.(Person).MyHead,FriendsName:x.(Person).FriendsName}
		return
	})
  
 
  
  //after the manipulate get the result as slice like in this wise :
  result:=Orca.From(myPersonList.ToSlice()).Skip(0).Take(1).ToSlice() 
  
  
  

```
