package set2

import (
	"bytes"
	"cryptopals/utils"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Implement PKCS#7 padding
func TestChallenge9(t *testing.T) {
	t.Run("padding", func(t *testing.T) {
		got := PKCS7([]byte("YELLOW SUBMARINE"), 20)
		expected := []byte("YELLOW SUBMARINE\x04\x04\x04\x04")

		if !bytes.Equal(got, expected) {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
	t.Run("padding entire block", func(t *testing.T) {
		got := PKCS7([]byte("YELLOW"), 6)
		expected := []byte("YELLOW\x06\x06\x06\x06\x06\x06")

		if !bytes.Equal(got, expected) {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
	t.Run("empty input", func(t *testing.T) {
		got := PKCS7(nil, 4)
		if got != nil {
			t.Error("expected an error")
		}
	})
	t.Run("padding reverse", func(t *testing.T) {
		got, _ := RemovePKCS7([]byte("YELLOW SUBMARINE\x04\x04\x04\x04"), 20)
		expected := []byte("YELLOW SUBMARINE")

		if !bytes.Equal(got, expected) {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
	t.Run("padding reverse entire block", func(t *testing.T) {
		got, _ := RemovePKCS7([]byte("YELLOW\x06\x06\x06\x06\x06\x06"), 6)
		expected := []byte("YELLOW")

		if !bytes.Equal(got, expected) {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
}

// Implement AES CBC
func TestChallenge10(t *testing.T) {
	t.Run("full circle", func(t *testing.T) {
		key := []byte("YELLOW SUBMARINE")
		iv := []byte("0000000000000000")

		input := []byte("something extremely random")

		encrypted := AESCBCEncrypt(input, key, iv)
		decrypt := AESCBCDecrypt(encrypted, key, iv)

		data, err := RemovePKCS7(decrypt, len(key))
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(input, data) {
			t.Errorf("got: %s, expected: %s", decrypt, input)
		}
	})

	t.Run("Make sure file ", func(t *testing.T) {
		key := []byte("YELLOW SUBMARINE")
		iv := bytes.Repeat([]byte("\x00"), len(key))

		file, err := utils.GetFile("https://cryptopals.com/static/challenge-data/10.txt")
		if err != nil {
			t.Error(err)
		}

		encodedCiphertext := strings.Replace(string(file), "\n", "", -1)
		rawCiphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
		if err != nil {
			t.Error(err)
		}

		decrypt := AESCBCDecrypt(rawCiphertext, key, iv)

		if !strings.Contains(string(decrypt), "Play that funky music") {
			t.Error("challenge 10 file didn't decrypt properly")
		}
	})
}

func TestChallenge11(t *testing.T) {
	t.Run("guess block cipher", func(t *testing.T) {
		input := []byte("things repeated things repeated things repeated")
		data, expected := OracleEncrypt(input)

		got := GuessBlockCipher(data)

		if got != expected {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
}

func TestChallenge12(t *testing.T) {
	t.Run("Guess key size", func(t *testing.T) {
		encryptor := NewEncryptor()

		guess := encryptor.KeySize()

		if guess != len(encryptor.key) {
			t.Errorf("got: %d, expected: %d", guess, len(encryptor.key))
		}
	})

	t.Run("Make sure we're using ECB", func(t *testing.T) {
		e := NewEncryptor()

		cipherText := e.Encrypt([]byte(strings.Repeat("A", e.KeySize()*3)))
		got := GuessBlockCipher(cipherText)

		if got != "ECB" {
			t.Errorf("got: %s, expected: ECB", got)
		}
	})

	t.Run("Retrieve the unkown", func(t *testing.T) {
		e := NewEncryptor()

		res := e.BruteForceSingleByte()
		if !strings.Contains(string(res), "Did you stop? No, I just drove by") {
			t.Errorf("got: %s", string(res))
		}
	})
}

func TestChallenge13(t *testing.T) {
	t.Run("decode URL", func(t *testing.T) {
		got := URLDecode("email=bar&uid=1234&role=user")
		expected := Profile{
			Email: "bar",
			UID:   "1234",
			Role:  "user",
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf("Expected (-) but got (+):\n%s", diff)
		}
	})
	t.Run("encode URL", func(t *testing.T) {
		got := URLEncode(Profile{
			Email: "bar",
			UID:   "1234",
			Role:  "user",
		})
		expected := "email=bar&uid=1234&role=user"
		if got != expected {
			t.Errorf("got: %s, expected: %s", got, expected)
		}
	})
	t.Run("profile for", func(t *testing.T) {
		got := URLDecode(ProfileFor("john.doe@test&people.com"))
		expected := Profile{
			Email: "john.doe@testpeople.com",
			Role:  "user",
		}

		if got.UID == "" {
			t.Errorf("Expected a UID but none was found: %v", got)
		}

		expected.UID = got.UID

		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf("Expected (-) but got (+):\n%s", diff)
		}
	})
	t.Run("profile encryptor round trip", func(t *testing.T) {
		e := NewProfileEncryptor()
		target_email := "test@people.com"
		cipher := e.Encrypt(target_email)
		data := URLDecode(e.Decrypt(cipher))

		if data.Email != target_email {
			t.Errorf("error encrypting/decrypting profile")
		}
	})

	// Given an email address, can we make our profile "token"
	// escalate from "user" to "admin" without knowing the encryption key?
	t.Run("make profile admin", func(t *testing.T) {
		e := NewProfileEncryptor()

		profile, err := e.Escalate("me@test.com")
		if err != nil {
			t.Error(err)
		}

		// very that the profile is now an admin
		data := URLDecode(e.Decrypt(profile))
		if data.Role != "admin" {
			t.Errorf("was expecting role: admin but got: %s", data.Role)
		}
	})
}

func TestChallenge14(t *testing.T) {
	t.Run("retrieve target", func(t *testing.T) {
		e := NewEncryptor()
		test := e.BruteForceSingleByteHarder()
		fmt.Println(test)
	})
}
