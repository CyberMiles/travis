package server

// Stop channel used to notify the server to shutdown
// Be set when the approved retire governance proposal has reached the aimed block height
var StopFlag = make(chan bool)
