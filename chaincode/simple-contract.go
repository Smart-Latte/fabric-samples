package main

import(
	"github.com/hyperldger/fabric-contract-api-go/contractapi"
)

func main(){
	simpleContract := new(SimpleContract)

	cc, err := contractapi.NewChainCod(simpleContract)

	if err != nil{
		panic(err.Error())
	}

	if err := cc.Start(); err != nil{
		panic(err.Error())
	}
}