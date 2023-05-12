package entities

const ETHEREUM = "eth"
const ARBITRUM = "arb"
const POLYGON = "matic"
const OPTIMISM = "op"

var SupportedChains = []string{ETHEREUM, ARBITRUM, POLYGON, OPTIMISM}

func ValidateChain(chain string) bool {
	switch chain {
	case ETHEREUM, ARBITRUM, POLYGON, OPTIMISM:
		return true
	default:
		return false
	}
}
