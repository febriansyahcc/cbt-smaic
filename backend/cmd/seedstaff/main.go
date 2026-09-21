// Command seedstaff memasukkan guru, mapel, kelas, dan penugasan kelas-mapel dari berkas
// data ke basis data PostgreSQL (DB_DSN). Bawaan adalah DRY-RUN; tambahkan --apply untuk menulis.
//
// Keluaran: laporan manusia ke stderr; CSV kredensial (username,nama_lengkap,password) ke
// stdout HANYA untuk akun baru dan HANYA setelah --apply ter-commit. Alihkan stdout ke berkas
// lokal di luar container, mis. `> seed-output/guru.csv`.
//
// Alat ini sengaja tidak memanggil InitDB, seed awal, backfill, maupun AutoMigrate.
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"cbt-backend/internal/repository"
	"cbt-backend/internal/staffseed"

	"gorm.io/gorm"
)

//go:embed data/staff_2026_2027.json
var defaultData []byte

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv, repository.Open))
}

// run berisi seluruh alur CLI; dipisahkan dari main agar dapat diuji. Mengembalikan exit code.
func run(args []string, stdout, stderr io.Writer, getenv func(string) string, open func(dsn string) (*gorm.DB, error)) int {
	fs := flag.NewFlagSet("seedstaff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "path berkas JSON data seed (bawaan: data yang di-embed di biner)")
	apply := fs.Bool("apply", false, "tulis perubahan ke basis data (bawaan: dry-run, tidak menulis apa pun)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "Galat: argumen tak dikenal: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	// Data.
	var src io.Reader = bytes.NewReader(defaultData)
	source := "data bawaan (embed)"
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			fmt.Fprintf(stderr, "Galat: gagal membuka berkas data: %v\n", err)
			return 1
		}
		defer f.Close()
		src = f
		source = *file
	}
	data, err := staffseed.LoadData(src)
	if err != nil {
		fmt.Fprintf(stderr, "Galat: %v\n", err)
		return 1
	}
	fmt.Fprintf(stderr, "Sumber data    : %s\n", source)

	// Koneksi. DSN dicetak dengan kata sandi disamarkan.
	dsn := getenv("DB_DSN")
	if dsn == "" {
		fmt.Fprintln(stderr, "Galat: env DB_DSN belum diisi")
		return 1
	}
	fmt.Fprintf(stderr, "Basis data     : %s\n", repository.RedactSecrets(dsn))
	db, err := open(dsn)
	if err != nil {
		fmt.Fprintf(stderr, "Galat: %s\n", repository.RedactSecrets(err.Error()))
		return 1
	}

	rep, err := staffseed.Run(db, data, *apply, nil, nil)
	if err != nil {
		fmt.Fprintf(stderr, "Galat: %s\nTidak ada perubahan yang tersimpan.\n", repository.RedactSecrets(err.Error()))
		return 1
	}
	if err := rep.WriteText(stderr); err != nil {
		fmt.Fprintf(stderr, "Peringatan: gagal menulis laporan: %v\n", err)
	}

	// Kredensial: hanya setelah commit berhasil dan hanya untuk akun baru.
	if rep.Committed && len(rep.Credentials) > 0 {
		if err := staffseed.WriteCredentialsCSV(stdout, rep.Credentials); err != nil {
			fmt.Fprintf(stderr, "Galat: perubahan SUDAH ter-commit tetapi gagal menulis CSV kredensial ke stdout: %v\n"+
				"Kata sandi awal tidak dapat dipulihkan; reset kata sandi akun baru lewat menu admin.\n", err)
			return 1
		}
	}
	return 0
}
