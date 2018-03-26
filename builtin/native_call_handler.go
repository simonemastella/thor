package builtin

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/vechain/thor/builtin/abi"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/vm"
)

func init() {
	methods := []*nativeMethod{
		Params.impl("nativeGetExecutor", ethparams.SloadGas, func(*env) ([]interface{}, error) {
			return []interface{}{Executor.Address}, nil
		}),
		Params.impl("nativeGet", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			var key common.Hash
			env.Args(&key)
			v := Params.WithState(env.State).Get(thor.Hash(key))
			return []interface{}{v}, nil
		}),
		Params.impl("nativeSet", ethparams.SstoreSetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				Key   common.Hash
				Value *big.Int
			}
			env.Args(&args)
			Params.WithState(env.State).Set(thor.Hash(args.Key), args.Value)
			return nil, nil
		}),

		Authority.impl("nativeGetExecutor", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			return []interface{}{Executor.Address}, nil
		}),
		Authority.impl("nativeAdd", ethparams.SstoreSetGas+ethparams.SstoreResetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				Signer   common.Address
				Endorsor common.Address
				Identity common.Hash
			}
			env.Args(&args)
			ok := Authority.WithState(env.State).Add(thor.Address(args.Signer), thor.Address(args.Endorsor), thor.Hash(args.Identity))
			return []interface{}{ok}, nil
		}),
		Authority.impl("nativeRemove", ethparams.SstoreClearGas, func(env *env) ([]interface{}, error) {
			var signer common.Address
			env.Args(&signer)
			ok := Authority.WithState(env.State).Remove(thor.Address(signer))
			return []interface{}{ok}, nil
		}),
		Authority.impl("nativeGet", ethparams.SloadGas*3, func(env *env) ([]interface{}, error) {
			var signer common.Address
			env.Args(&signer)
			p := Authority.WithState(env.State).Get(thor.Address(signer))
			return []interface{}{!p.IsEmpty(), p.Endorsor, p.Identity, p.Active}, nil
		}),
		Authority.impl("nativeFirst", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			signer := Authority.WithState(env.State).First()
			return []interface{}{signer}, nil
		}),
		Authority.impl("nativeNext", ethparams.SloadGas*4, func(env *env) ([]interface{}, error) {
			var signer common.Address
			env.Args(&signer)
			p := Authority.WithState(env.State).Get(thor.Address(signer))
			var next thor.Address
			if p.Next != nil {
				next = *p.Next
			}
			return []interface{}{next}, nil
		}),

		Energy.impl("nativeGetExecutor", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			return []interface{}{Executor.Address}, nil
		}),
		Energy.impl("nativeGetTotalSupply", 3000, func(env *env) ([]interface{}, error) {
			supply := Energy.WithState(env.State).GetTotalSupply(env.VMContext.Time)
			return []interface{}{supply}, nil
		}),
		Energy.impl("nativeGetTotalBurned", ethparams.SloadGas*2, func(env *env) ([]interface{}, error) {
			burned := Energy.WithState(env.State).GetTotalBurned()
			return []interface{}{burned}, nil
		}),
		Energy.impl("nativeGetBalance", 2000, func(env *env) ([]interface{}, error) {
			var addr common.Address
			env.Args(&addr)
			bal := Energy.WithState(env.State).GetBalance(env.VMContext.Time, thor.Address(addr))
			return []interface{}{bal}, nil
		}),
		Energy.impl("nativeAddBalance", ethparams.SstoreSetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				Addr   common.Address
				Amount *big.Int
			}
			env.Args(&args)
			Energy.WithState(env.State).AddBalance(env.VMContext.Time, thor.Address(args.Addr), args.Amount)
			return nil, nil
		}),
		Energy.impl("nativeSubBalance", ethparams.SstoreResetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				Addr   common.Address
				Amount *big.Int
			}
			env.Args(&args)
			ok := Energy.WithState(env.State).SubBalance(env.VMContext.Time, thor.Address(args.Addr), args.Amount)
			return []interface{}{ok}, nil
		}),
		Energy.impl("nativeAdjustGrowthRate", ethparams.SstoreSetGas, func(env *env) ([]interface{}, error) {
			var rate *big.Int
			env.Args(&rate)
			Energy.WithState(env.State).AdjustGrowthRate(env.VMContext.Time, rate)
			return nil, nil
		}),
		Energy.impl("nativeApproveConsumption", ethparams.SstoreSetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				ContractAddr common.Address
				Caller       common.Address
				Credit       *big.Int
				RecoveryRate *big.Int
				Expiration   uint64
			}
			env.Args(&args)
			Energy.WithState(env.State).ApproveConsumption(env.VMContext.Time,
				thor.Address(args.ContractAddr), thor.Address(args.Caller), args.Credit, args.RecoveryRate, args.Expiration)
			return nil, nil
		}),
		Energy.impl("nativeGetConsumptionAllowance", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			var args struct {
				ContractAddr common.Address
				Caller       common.Address
			}
			env.Args(&args)
			remained := Energy.WithState(env.State).GetConsumptionAllowance(env.VMContext.Time,
				thor.Address(args.ContractAddr), thor.Address(args.Caller))
			return []interface{}{remained}, nil
		}),
		Energy.impl("nativeSetSupplier", ethparams.SstoreSetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				ContractAddr common.Address
				Supplier     common.Address
				Agreed       bool
			}
			env.Args(&args)
			Energy.WithState(env.State).SetSupplier(thor.Address(args.ContractAddr), thor.Address(args.Supplier), args.Agreed)
			return nil, nil
		}),
		Energy.impl("nativeGetSupplier", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			var contractAddr common.Address
			env.Args(&contractAddr)
			supplier, ok := Energy.WithState(env.State).GetSupplier(thor.Address(contractAddr))
			return []interface{}{supplier, ok}, nil
		}),
		Energy.impl("nativeSetContractMaster", ethparams.SstoreResetGas, func(env *env) ([]interface{}, error) {
			var args struct {
				ContractAddr common.Address
				Master       common.Address
			}
			env.Args(&args)
			Energy.WithState(env.State).SetContractMaster(thor.Address(args.ContractAddr), thor.Address(args.Master))
			return nil, nil
		}),
		Energy.impl("nativeGetContractMaster", ethparams.SloadGas, func(env *env) ([]interface{}, error) {
			var contractAddr common.Address
			env.Args(&contractAddr)
			master := Energy.WithState(env.State).GetContractMaster(thor.Address(contractAddr))
			return []interface{}{master}, nil
		}),
	}

	for _, method := range methods {
		methodMap[methodKey{
			method.addr, method.methodCodec.ID(),
		}] = method
	}
}

type methodKey struct {
	thor.Address
	abi.MethodID
}

var methodMap = make(map[methodKey]*nativeMethod)

// HandleNativeCall entry of native methods implementaion.
func HandleNativeCall(state *state.State, vmCtx *vm.Context, to thor.Address, input []byte) func(useGas func(gas uint64) bool, caller thor.Address) ([]byte, error) {
	if len(input) < 4 {
		return nil
	}
	var methodID abi.MethodID
	copy(methodID[:], input)

	method := methodMap[methodKey{to, methodID}]
	if method == nil {
		return nil
	}

	return func(useGas func(gas uint64) bool, caller thor.Address) ([]byte, error) {
		return method.Call(state, vmCtx, caller, useGas, input)
	}
}
