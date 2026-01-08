package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alapierre/ksef-qr-generator/version"

	"github.com/akamensky/argparse"
	"github.com/alapierre/go-ksef-client/ksef"
	"github.com/alapierre/go-ksef-client/ksef/keys"
	"github.com/alapierre/go-ksef-client/ksef/qr"
	"github.com/alapierre/go-ksef-client/png"
	"golang.org/x/term"
)

func main() {

	parser := argparse.NewParser("qr-gen", "KSeF QR Code II Generator, wersja "+version.Version)

	cert := parser.File("c", "cert", os.O_RDONLY, 0, &argparse.Options{Required: true, Help: "Plik certyfikatu KSeF"})
	key := parser.File("k", "key", os.O_RDONLY, 0, &argparse.Options{Required: true, Help: "Plik klucza prywatnego"})
	env := parser.String("e", "env", &argparse.Options{Required: false, Help: "Środowisko (test, demo, prod)", Default: "test"})
	seller := parser.String("s", "seller-nip", &argparse.Options{Required: false, Help: "NIP sprzedawcy, jeśli inny niż NIP wystawcy faktury"})
	nip := parser.String("n", "context-nip", &argparse.Options{Required: true, Help: "NIP wystawcy faktury (kontekstu KSeF)"})
	out := parser.String("o", "out", &argparse.Options{Required: false, Help: "ścieżka bazowa do zapisu QR Code, brak oznacza zapis w bieżącym katalogu"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	pass := os.Getenv("KSEF_KEY_PASSWORD")
	if pass == "" {
		pass, err = ReadPassword("Podaj hasło do klucza prywatnego: ")
		if err != nil {
			fmt.Println("Błąd odczytu hasła: ", err)
			os.Exit(1)
		}
	}

	cBytes, err := io.ReadAll(cert)
	if err != nil {
		fmt.Println("Błąd odczytu certyfikatu: ", err)
		os.Exit(1)
	}
	c, err := qr.LoadCertificate(cBytes)
	if err != nil {
		fmt.Println("Błąd odczytu certyfikatu: ", err)
		os.Exit(1)
	}

	serial, err := qr.ExtractCertSerial(c)
	if err != nil {
		fmt.Println("Błąd odczytu numeru seryjnego certyfikatu: ", err)
		os.Exit(1)
	}

	kBytes, err := io.ReadAll(key)
	if err != nil {
		fmt.Println("Błąd odczytu klucza prywatnego: ", err)
		os.Exit(1)
	}

	k, err := keys.LoadEncryptedPKCS8SignerFromPEM(kBytes, []byte(pass))
	if err != nil {
		fmt.Println("Błąd odczytu klucza prywatnego: ", err)
		os.Exit(1)
	}

	var e ksef.Environment
	err = e.UnmarshalText([]byte(*env))
	if err != nil {
		fmt.Println("Błędne środowisko KSeF ", err)
		os.Exit(1)
	}

	if *seller == "" {
		seller = nip
	}

	sum := sha256.Sum256([]byte("KSeF is dead, baby, KSeF is dead..."))
	url, err := qr.GenerateCertificateVerificationLink(
		e,
		qr.CtxNip,
		*nip,
		*seller,
		serial,
		k,
		sum[:],
	)

	if err != nil {
		fmt.Println("Błąd generowania linku: ", err)
		os.Exit(1)
	}

	fmt.Printf("Link wygenerowany dla NIP sprzedawcy %s, kontekst KSeF (wystawca) %s, środowisko: %s nr seryjny certyfikatu %s\n", *seller, *nip, e.Name(), serial)
	fmt.Println(url)

	img, err := png.Qr(url)
	if err != nil {
		return
	}
	outPath := filepath.Join(*out, fmt.Sprintf("%s_qr2.png", *nip))
	err = os.WriteFile(outPath, img, 0o644)
	if err != nil {
		fmt.Println("Błąd zapisu QR kodu: ", err)
		os.Exit(1)
	}

	fmt.Printf("QR kod saved %s. Your visualization will be formally correct. Substantively… maybe. 😉 Happy KSeFing!\n", outPath)
}

func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println() // Add a newline after the password entry
	return strings.TrimSpace(string(bytePassword)), nil
}
