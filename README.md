gophervalidate
===========

Basic validation library which allows you to place expensive validation functions in goroutines to execute concurrently. Written in part to learn Go.

Has no external dependencies


#### Installation
Ensure Go is installed on your computer.
Run the standard go get as so:

	go get github.com/willtrking/gophervalidate


#### Example
```go

func RealHeavyValidator(userID int, v *gophervalidate.Validator) {

  // Do a ton of heavy work here
  if(userID == 1){
    v.RecordOK("id")
  }else{
    v.RecordError("id","Unknown user ID %d", userID)
  }

}

func InsaneValidator(userID int, v *gophervalidate.Validator) {

  //In this case we want to try to grab a result from another key
  //So we need to wait until its processed
  result := v.WaitForKey("age")

  if result.isError {

    v.RecordError("secondID",result.msg)

  }else{

    // Do a ton of heavy work here
    if(userID == 1){
      v.RecordOK("secondID")
    }else{
      v.RecordError("secondID","Unknown user ID %d", userID)
    }

  }

}

userID := 1
userAge := 18

validator = gophervalidate.MakeValidator()
//We need to make sure we add how many validators will be userID

validator.AddValidators(2)
validator.CheckBool("age", 18 > 21, "You're underage, must be at least %d!", 21)
go RealHeavyValidator(userID, validator)

otherUserId := 2
//Oops forgot one! Can always increment, incrementation is automic
validator.AddValidators(1)
go RealHeavyValidator(otherUserId, validator)

//The above 3 validators execute concurrently

//Wait for all of our validation to finish and get all of the errors
errMap := validator.ValidateAndClose()

//Error map is a map in the form *map[string][]string
//Map keys are the first values passed to CheckBool / RecordOK / RecordErr
//Values are lists of error strings
```
