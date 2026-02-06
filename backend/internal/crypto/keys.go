package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP recommendations)
const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
)

// BIP39 word list (first 256 words for simplicity - full list has 2048)
// In production, use a proper BIP39 library
var wordList = []string{
	"abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract",
	"absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid",
	"acoustic", "acquire", "across", "act", "action", "actor", "actress", "actual",
	"adapt", "add", "addict", "address", "adjust", "admit", "adult", "advance",
	"advice", "aerobic", "affair", "afford", "afraid", "again", "age", "agent",
	"agree", "ahead", "aim", "air", "airport", "aisle", "alarm", "album",
	"alcohol", "alert", "alien", "all", "alley", "allow", "almost", "alone",
	"alpha", "already", "also", "alter", "always", "amateur", "amazing", "among",
	"amount", "amused", "analyst", "anchor", "ancient", "anger", "angle", "angry",
	"animal", "ankle", "announce", "annual", "another", "answer", "antenna", "antique",
	"anxiety", "any", "apart", "apology", "appear", "apple", "approve", "april",
	"arch", "arctic", "area", "arena", "argue", "arm", "armed", "armor",
	"army", "around", "arrange", "arrest", "arrive", "arrow", "art", "artefact",
	"artist", "artwork", "ask", "aspect", "assault", "asset", "assist", "assume",
	"asthma", "athlete", "atom", "attack", "attend", "attitude", "attract", "auction",
	"audit", "august", "aunt", "author", "auto", "autumn", "average", "avocado",
	"avoid", "awake", "aware", "away", "awesome", "awful", "awkward", "axis",
	"baby", "bachelor", "bacon", "badge", "bag", "balance", "balcony", "ball",
	"bamboo", "banana", "banner", "bar", "barely", "bargain", "barrel", "base",
	"basic", "basket", "battle", "beach", "bean", "beauty", "because", "become",
	"beef", "before", "begin", "behave", "behind", "believe", "below", "belt",
	"bench", "benefit", "best", "betray", "better", "between", "beyond", "bicycle",
	"bid", "bike", "bind", "biology", "bird", "birth", "bitter", "black",
	"blade", "blame", "blanket", "blast", "bleak", "bless", "blind", "blood",
	"blossom", "blouse", "blue", "blur", "blush", "board", "boat", "body",
	"boil", "bomb", "bone", "bonus", "book", "boost", "border", "boring",
	"borrow", "boss", "bottom", "bounce", "box", "boy", "bracket", "brain",
	"brand", "brass", "brave", "bread", "breeze", "brick", "bridge", "brief",
	"bright", "bring", "brisk", "broccoli", "broken", "bronze", "broom", "brother",
	"brown", "brush", "bubble", "buddy", "budget", "buffalo", "build", "bulb",
	"bulk", "bullet", "bundle", "bunker", "burden", "burger", "burst", "bus",
	"business", "busy", "butter", "buyer", "buzz", "cabbage", "cabin", "cable",
}

// DeriveKeyFromSecret derives a 256-bit key from a secret using Argon2id
func DeriveKeyFromSecret(secret, salt []byte) []byte {
	return argon2.IDKey(secret, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

// DeriveKeyFromPassphrase derives a key from a recovery passphrase
func DeriveKeyFromPassphrase(passphrase string, salt []byte) []byte {
	// Normalize passphrase
	words := strings.Fields(strings.ToLower(strings.TrimSpace(passphrase)))
	normalized := strings.Join(words, " ")
	return DeriveKeyFromSecret([]byte(normalized), salt)
}

// GenerateRecoveryPhrase generates a BIP39-style recovery phrase
func GenerateRecoveryPhrase() (string, error) {
	// Generate 16 bytes of entropy (128 bits = 12 words)
	entropy := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, entropy); err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Convert entropy to word indices
	// Each word represents ~10.67 bits of entropy
	words := make([]string, 12)
	for i := 0; i < 12; i++ {
		// Use entropy bytes to select words
		idx := int(entropy[i%16]) ^ int(entropy[(i+1)%16])
		words[i] = wordList[idx%len(wordList)]
	}

	return strings.Join(words, " "), nil
}

// HashRecoveryPhrase creates a hash of the recovery phrase for verification
func HashRecoveryPhrase(passphrase string) []byte {
	normalized := strings.ToLower(strings.TrimSpace(passphrase))
	hash := sha256.Sum256([]byte(normalized))
	return hash[:]
}

// VerifyRecoveryPhrase checks if a passphrase matches the stored hash
func VerifyRecoveryPhrase(passphrase string, storedHash []byte) bool {
	hash := HashRecoveryPhrase(passphrase)
	if len(hash) != len(storedHash) {
		return false
	}
	// Constant-time comparison
	var diff byte
	for i := range hash {
		diff |= hash[i] ^ storedHash[i]
	}
	return diff == 0
}

// EncryptDEK encrypts the DEK with the KEK (derived from WebAuthn PRF or passphrase)
func EncryptDEK(dek, kek []byte) ([]byte, error) {
	return Encrypt(dek, kek)
}

// DecryptDEK decrypts the DEK using the KEK
func DecryptDEK(encryptedDEK, kek []byte) ([]byte, error) {
	return Decrypt(encryptedDEK, kek)
}

// GenerateRandomHex generates a random hex string of the specified byte length
func GenerateRandomHex(byteLen int) (string, error) {
	bytes := make([]byte, byteLen)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
