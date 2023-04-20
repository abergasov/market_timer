package entities

const ETH = "eth"

func ValidateChain(chain string) bool {
	switch chain {
	case ETH:
		return true
	default:
		return false
	}
}
