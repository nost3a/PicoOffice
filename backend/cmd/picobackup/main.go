// picobackup: standalone backup subcommand; SQLite snapshot + storage -> tar.gz.
//
// usage:
//
//	picobackup --db ./picooffice.db --storage ../storage --out ./backup-20261007.tar.gz
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nost3a/PicoOffice/internal/service"
)

func main() {
	dbPath := flag.String("db", "./picooffice.db", "sqlite db 文件路径")
	storage := flag.String("storage", "../storage", "storage 目录")
	out := flag.String("out", "", "输出 tar.gz 路径（必填）")
	flag.Parse()

	if *out == "" {
		flag.Usage()
		log.Fatal("--out is required")
	}

	if err := service.Backup(*dbPath, *storage, *out); err != nil {
		log.Fatalf("backup failed: %v", err)
	}
	if st, err := os.Stat(*out); err == nil {
		fmt.Printf("backup written: %s (%d bytes)\n", *out, st.Size())
	} else {
		fmt.Println("backup written:", *out)
	}
}
