package gophervalidate

import (
	"fmt"
	"sync/atomic"
)

type Validator struct {
	chanSize      uint32
	validatorChan chan ValidatorMessage
}

type ValidatorMessage struct {
	key     string
	msg     string
	isError bool
}

//Create our validator, create a channel to take errors
func NewValidator() *Validator {

	v := new(Validator)
	v.validatorChan = make(chan ValidatorMessage)
	v.chanSize = 0

	return v
}

//Increment our channel size
//Required to properly wait for all validators to complete
func (v *Validator) AddValidators(addv uint32) {
	atomic.AddUint32(&v.chanSize, addv)
}

//Record a message
func (v *Validator) RecordMessage(key string, msg string, err bool) {
	v.validatorChan <- ValidatorMessage{key, msg, err}
}

//Record an error
func (v *Validator) RecordError(key string, errMsg string, a ...interface{}) {
	go v.RecordMessage(key, fmt.Sprintf(errMsg, a...), true)
}

//Record OK
func (v *Validator) RecordOK(key string) {
	go v.RecordMessage(key, "", false)
}

//Basic bool validation
func (v *Validator) CheckBool(key string, c bool, errMsg string, a ...interface{}) {
	if !c {
		v.RecordError(key, errMsg, a...)
	} else {
		v.RecordOK(key)
	}
}

//Wait for our validation to be done, and consume the channel
func (v *Validator) Validate() map[string][]string {

	errMap := make(map[string][]string)

	for i := uint32(0); i < v.chanSize; i++ {

		msg := <-v.validatorChan

		if msg.isError {
			if val, ok := errMap[msg.key]; ok {
				errMap[msg.key] = append(val, msg.msg)
			} else {
				errMap[msg.key] = []string{msg.msg}
			}
		}
	}

	return errMap

}

//Waits for a key to be validated, and returns a copy of the result
//Result will remain in validator channel
func (v *Validator) WaitForKey(key string) *ValidatorMessage {

	var res ValidatorMessage

	var pulledResults []ValidatorMessage

	for i := uint32(0); i < v.chanSize; i++ {
		msg := <-v.validatorChan

		pulledResults = append(pulledResults, msg)

		if msg.key == key {
			res = msg
			break
		}

	}

	//Re-record all the messages we pulled out
	for _, pulled := range pulledResults {
		go v.RecordMessage(pulled.key, pulled.msg, pulled.isError)
	}

	return &res

}

//Same as validate, but also close
func (v *Validator) ValidateAndClose() map[string][]string {
	//Close out when we're done
	defer v.Close()
	return v.Validate()
}

//Close validator
func (v *Validator) Close() {
	if v.validatorChan != nil {
		close(v.validatorChan)
	}
	*v = Validator{}
}

//Reset validator to state from NewValidator
func (v *Validator) Reset() {
	v.Close()
	v = NewValidator()
}
