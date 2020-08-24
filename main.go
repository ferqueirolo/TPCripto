package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"log"
	"math/big"
)

/*
	TRABAJO PRACTICO CRIPTROGRAFIA
	TEMA: ETHEREUM - TRANSACCIONES

*/

func main() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(errors.Errorf("ERROR: %v", r))
			fmt.Printf("\nPress 'Enter' to finish...")
			fmt.Scanln()
		}
	}()
	/// - - - -  - - - -  - - - -  SETUP - - - -  - - - -  - - - -  - - - -
	ctx := context.Background()
	// - - - - CLIENTE A
	// Creamos la clave privada del cliente A
	privateKeyA, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	// Creamos la clave publica del cliente A
	publicKeyA := bind.NewKeyedTransactor(privateKeyA)

	// - - - - CLIENTE B
	// Creamos la clave privada del cliente B
	privateKeyB, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	// Creamos la clave publica del cliente B
	publicKeyB := bind.NewKeyedTransactor(privateKeyB)

	// BALANCE DE LA CUENTA
	// La unidad más pequeña de Ether -wei-. Se necesitan muchos wei para hacer un Ether. 10^18, para ser exactos.
	balance := new(big.Int)
	balance.SetString("10000000", 10)

	// INICIALIZAMOS LAS CUENTAS DE LOS CLIENTES
	genesisChain := map[common.Address]core.GenesisAccount{
		// cuenta cliente B
		publicKeyA.From: {
			Balance: balance,
		},
		// cuenta cliente B
		publicKeyB.From: {
			Balance: balance,
		},
	}
	fmt.Println("CLIENTES INICIALES")
	printBalance("A", balance)
	fmt.Println("Clave publica A: " + publicKeyA.From.String())
	fmt.Println()
	printBalance("B", balance)
	fmt.Println("Clave publica B: " + publicKeyB.From.String())
	fmt.Println("- - - - - - - - - - -")

	// CONFIGURAMOS LA RED SIMULADA DE ETHEREUM
	blockGasLimit := uint64(4712388)
	backend := backends.NewSimulatedBackend(genesisChain, blockGasLimit)

	/// - - - - - - - - - - - -  TRANSACCIONES - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

	/*	- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
			TRANSACCIÓN Nº1
			CLIENTE A ---> CLIENTE B
			El cliente A le enviara la suma de 1000 WEI al cliente B
		 	- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	*/

	// NUMERO DE TRANSACCIONES ASOCIADAS A UNA CUENTA
	nonce, err := backend.PendingNonceAt(ctx, publicKeyA.From)
	if err != nil {
		log.Fatal(err)
	}

	var data []byte

	// Monto de la primera transacción
	a := getAmount()
	amount := big.NewInt(a)

	// Gas limite de la primera transacción - MINIMO COSTO DE TRANSACCION ES 21000 WEI
	limit := getGasLimit()
	gasLimit := uint64(limit)

	// Multiplo del precio de la gas de la primera transacción
	gasPrice := big.NewInt(1)

	// El cliente A genera la primera transacción con la clave publica del cliente B
	trx1AtoB := types.NewTransaction(nonce, publicKeyB.From, amount, gasLimit, gasPrice, data)

	chainID := big.NewInt(1337)

	// El cliente A firma la transacción a enviar al cliente B
	signedATx, err := types.SignTx(trx1AtoB, types.NewEIP155Signer(chainID), privateKeyA)
	if err != nil {
		log.Fatal(err)
	}

	// Se envia la transacción
	if err = backend.SendTransaction(ctx, signedATx); err != nil {
		fmt.Println("ERROR Gas insuficiente")
		return
	}

	// Confirmamos las transacciones para que se impacten en los balances
	backend.Commit()

	// Extraemos los balances de los clientes
	balance1A := extractBalance(err, backend, publicKeyA)
	balance1B := extractBalance(err, backend, publicKeyB)

	block1 := backend.Blockchain().CurrentBlock()
	trx1 := block1.Transactions()[0]
	fmt.Printf("A - - - %v - - - > B\n\n", trx1.Value())
	printTrxInfo(trx1)
	printBalance("A", extractBalance(err, backend, publicKeyA))
	printBalance("B", extractBalance(err, backend, publicKeyB))

	err = backend.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("- - - - - - - - - - -")
	fmt.Println()

	/*  	- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	TRANSACCIÓN Nº2
	CLIENTE B ---> CLIENTE A
	El cliente B le enviara la suma de 2000 WEI al cliente A
	 - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	*/

	genesisChain2 := map[common.Address]core.GenesisAccount{
		// cuenta cliente B
		publicKeyA.From: {
			Balance: balance1A,
		},
		// cuenta cliente B
		publicKeyB.From: {
			Balance: balance1B,
		},
	}

	// NUMERO DE TRANSACCIONES ASOCIADAS A UNA CUENTA
	nonce, err = backend.PendingNonceAt(ctx, publicKeyB.From)
	if err != nil {
		log.Fatal(err)
	}

	// CONFIGURAMOS LA RED SIMULADA DE ETHEREUM
	backend = backends.NewSimulatedBackend(genesisChain2, blockGasLimit)

	// Monto de la segunda transacción
	a = getAmount()
	amount = big.NewInt(a)

	// Gas limite de la segunda transacción - MINIMO COSTO DE TRANSACCION ES 21000 WEI
	limit = getGasLimit()
	gasLimit = uint64(limit)

	// Multiplo del precio de la gas de la segunda transacción
	gasPrice = big.NewInt(1)

	// El cliente B genera la segunda transacción con la clave publica del cliente A
	trx1BtoA := types.NewTransaction(nonce, publicKeyA.From, amount, gasLimit, gasPrice, data)

	// El cliente B firma la transacción a enviar al cliente A
	signedBTx, err := types.SignTx(trx1BtoA, types.NewEIP155Signer(chainID), privateKeyB)
	if err != nil {
		log.Fatal(err)
	}

	// Se envia la transacción
	err = backend.SendTransaction(ctx, signedBTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Confirmamos las transacciones para que se impacten en los balances
	backend.Commit()

	// Se imprime el balance de la transaccion
	trx2 := backend.Blockchain().CurrentBlock().Transactions()[0]
	fmt.Printf("B - - - %v - - - > A\n", trx2.Value())

	// Imprimimos los detalles de la segunda transaccion
	printTrxInfo(trx2)

	// Imprimimos los balances
	printBalance("A", extractBalance(err, backend, publicKeyA))
	printBalance("B", extractBalance(err, backend, publicKeyB))

	fmt.Printf("\nPress 'Enter' to finish...")
	fmt.Scanln()
}

// OBTIENE DEL INPUT EL GAS LIMIT
func getGasLimit() int64 {
	fmt.Print("Gas limit [Minimo 21000]: ")
	var a int64
	fmt.Scanln(&a)
	return a
}

// OBTIENE DEL INPUT EL AMOUNT
func getAmount() int64 {
	fmt.Print("Monto transaccion: ")
	var a int64
	fmt.Scanln(&a)
	return a
}

// PRINTEA LA INFORMACIÓN DE UNA TRANSACCION
func printTrxInfo(trx *types.Transaction) {
	fmt.Printf("Amount: %v\n", trx.Value())
	fmt.Printf("GasPrice: %v\n", trx.GasPrice())
	fmt.Printf("Gas: %v\n", trx.Gas())
	fmt.Printf("Cost: %v + (%v * %v) = %v\n", trx.Value(), trx.Gas(), trx.GasPrice(),
		trx.Cost())
	fmt.Printf("To (public key): %v\n", trx.To().String())
	fmt.Println()
}

// EXTRAE EL BALANCE DE UNA RED SIMULADA PREGUNTANDO POR EL PRIMER BLOQUE Y POR CLAVE PUBLICA
func extractBalance(err error, backend *backends.SimulatedBackend, publicKey *bind.TransactOpts) *big.Int {
	balance, err := backend.BalanceAt(context.Background(), publicKey.From, big.NewInt(1))
	if err != nil {
		panic(err)
	}

	return balance
}

// PRINTEA EL BALANCE
func printBalance(cliente string, balance *big.Int) {
	fmt.Printf("BALANCE CLIENTE: "+cliente+" : %10v\n", balance.String())
}
