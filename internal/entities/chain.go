package entities

const ETH = "eth"
const ARB = "arb"
const MATIC = "matic"

func ValidateChain(chain string) bool {
	switch chain {
	case ETH, ARB, MATIC:
		return true
	default:
		return false
	}
}
