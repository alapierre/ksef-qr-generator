package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alapierre/ksef-qr-generator/bmp"
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
	out := parser.String("o", "out", &argparse.Options{Required: false, Help: "Ścieżka bazowa do zapisu QR Code, brak oznacza zapis w bieżącym katalogu"})
	red := parser.String("r", "redirect", &argparse.Options{Required: false, Help: "Ścieżka do pliku, w którym mają być zapisywane podpisane linki (dopisywanie linia po linii)"})

	format := parser.String("f", "format", &argparse.Options{Required: false, Help: "Format wyjściowy obrazka (png lub bmp)", Default: "png"})

	inPath := parser.String("i", "in", &argparse.Options{
		Required: false,
		Help:     "Opcjonalnie, plik XML faktury (lub inny, dowolny który ma zostać 'podpisany'), niezależnie od tego co podpiszemy, kod QR II i tak będzie ważny. Użyj '-' aby czytać ze stdin",
	})

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

	var sum [32]byte
	switch strings.TrimSpace(*inPath) {
	case "":
		fmt.Println("Używam domyślnej treści zamiast realnej faktury - bez obaw, kod QR i tak będzie ważny")
		sum = sha256.Sum256([]byte("KSeF is dead, baby, KSeF is dead..."))
	case "-":
		sum = shaOrDie(os.Stdin)
	default:
		f, err := os.Open(*inPath)
		if err != nil {
			fmt.Println("Błąd otwarcia pliku do podpisania: ", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		sum = shaOrDie(f)
	}

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
	if *red != "" {
		err := appendLine(*red, fmt.Sprintf("NIP sprzedawcy %s, kontekst KSeF (wystawca) %s, nr seryjny certyfikatu: %s, link: %s", *nip, *seller, serial, url))
		if err != nil {
			fmt.Printf("błąd zapisu informacji wyjściowej do pliku %s: %s\n", *red, err)
		}
	}

	var img []byte
	var ext string
	if *format == "png" {
		img, err = png.Qr(url)
		ext = "png"
	} else {
		img, err = bmp.Qr(url)
		ext = "bmp"
	}

	if err != nil {
		return
	}
	outPath := filepath.Join(*out, fmt.Sprintf("%s_qr2.%s", *nip, ext))
	err = os.WriteFile(outPath, img, 0o644)
	if err != nil {
		fmt.Println("Błąd zapisu QR kodu: ", err)
		os.Exit(1)
	}

	fmt.Printf("QR kod saved %s. Your visualization will be formally correct. Substantively… maybe. 😉 Happy KSeFing!\n", outPath)
}

func shaOrDie(r io.Reader) [32]byte {
	var sum [32]byte

	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		fmt.Println("Błąd liczenia sha z zawartości pliku do podpisu", err)
		os.Exit(1)
	}
	copy(sum[:], h.Sum(nil))

	return sum
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

func appendLine(path, line string) error {

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, line)
	return err
}
