package util

// Package util
// provides small, stateless utility functions shared across the application.
// Functions here have no dependencies on other internal packages.

import (
	"fmt"
	"math/rand/v2"
)

// adjectives and nouns form the human-readable component of a deployment slug.
// the wordlist is intentionally short for now. The uniqueness comes from the UUID suffix,
// not from wordlist size. words are chosen to be unambiguous when spoken aloud
// (no homophones like "blue/blew") and safe in a professional context.
// Can be swapped out for a slug generation library in the future.
// TODO: feature to allow users to customize the slug with their input text (with validation)
var adjectives = []string{
	"amber", "azure", "bold", "calm", "cedar", "clean", "clear",
	"crisp", "dawn", "dusk", "emerald", "fair", "firm", "fleet",
	"frost", "gold", "grand", "green", "grey", "iron", "jade",
	"keen", "lark", "lean", "light", "lunar", "maple", "mist",
	"noble", "north", "oak", "onyx", "open", "peak", "pine",
	"plain", "prime", "quick", "quiet", "rapid", "regal", "ridge",
	"river", "rose", "ruby", "sage", "sand", "sharp", "shore",
	"silk", "silver", "slate", "solar", "solid", "stark", "steel",
	"stone", "storm", "swift", "teal", "terra", "tidal", "true",
	"vale", "vast", "warm", "white", "wild", "wind",
}

var nouns = []string{
	"arc", "bay", "beam", "bird", "blade", "bloom", "bolt", "bond",
	"brook", "cliff", "cloud", "coast", "core", "crest", "crow",
	"dale", "dawn", "delta", "dune", "dust", "echo", "edge", "fern",
	"field", "flame", "flare", "fleet", "flow", "fog", "ford",
	"forge", "fox", "frost", "gale", "gate", "glen", "grove", "gust",
	"hawk", "hill", "horizon", "isle", "keep", "lake", "lark", "leaf",
	"light", "line", "lynx", "mast", "mesa", "mill", "mist", "moon",
	"moss", "mount", "node", "ore", "path", "peak", "pine", "plain",
	"pond", "pool", "port", "pulse", "ridge", "rift", "rise", "river",
	"rock", "root", "run", "sand", "seed", "shore", "sky", "slope",
	"snow", "sol", "spark", "spire", "spring", "star", "stem", "step",
	"stone", "stream", "sun", "surf", "surge", "tide", "trail", "tree",
	"vale", "veil", "vine", "wake", "wave", "wind", "wing", "wood",
}

// GenerateSlug returns a URL-safe slug in the format "adjective-noun-xxxx"
// where xxxx is a 4-character random hex suffix (for total uniqueness)
// the suffix provides enough entropy to make collisions statistically negligible (4^16 = 65536 possible suffixes)
// on a single-node deployment with a small number of total deployments.
// example output: "amber-ridge-3f9a", "swift-hawk-c142"
func GenerateSlug() string {
	// rand.IntN() generates a random integer in the range [0, n) where n is exclusive
	randomNumberForAdjective := rand.IntN(len(adjectives))
	adjective := adjectives[randomNumberForAdjective] // adjective at random index generated above

	noun := nouns[rand.IntN(len(nouns))]

	// rand.Uint32() generates a random 32-bit unsigned integer. (unsigned = no neg or pos)
	randomNumber := rand.Uint32()

	// masking with 0xFFFF isolates the lower 16 bits, giving a value in [0, 65535].
	maskedNumber := randomNumber & 0xFFFF

	// formatted as %04x this produces exactly 4 lowercase hex characters
	uuidSuffix := fmt.Sprintf("%04x", maskedNumber)

	return fmt.Sprintf("%s-%s-%s", adjective, noun, uuidSuffix)
}
