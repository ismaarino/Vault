package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"
	"github.com/manucorporat/try"
)

var (
	sep            = "~"
	initialVector  = "1234567890123456"
	vault_cipher   cipher.Block
	vault_filepath = "./data"
	vault_map      map[string]string
	main_w         *astilectron.Window
	login_w        *astilectron.Window
)

func showNotification(a *astilectron.Astilectron, title string, body string) {
	var n = a.NewNotification(&astilectron.NotificationOptions{
		Body:             body,
		HasReply:         astikit.BoolPtr(false),
		Icon:             "/path/to/icon",
		ReplyPlaceholder: "",
		Title:            title,
	})
	n.Create()
	n.Show()
}

func loadData(path string) {
	dat, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("1")
		vault_map["Default.com"+sep+"user"] = encryptAES(vault_cipher, "default")
		fmt.Println("1.5")
		storeData(vault_map, vault_filepath)
	} else {
		b := new(bytes.Buffer)
		b.Write(dat)
		d := gob.NewDecoder(b)

		// Decoding the serialized data
		err = d.Decode(&vault_map)
		if err != nil {
			panic(err)
		}
		aux_data := map[string]string{}
		for k, v := range vault_map {
			fmt.Println("loaded " + k)
			aux_data[decryptAES(vault_cipher, k)] = v
			delete(vault_map, k)
		}
		vault_map = aux_data
	}
}

func storeData(data map[string]string, path string) {
	fmt.Println("2")
	aux_data := map[string]string{}
	for k, v := range data {
		aux_data[encryptAES(vault_cipher, k)] = v
	}

	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(aux_data)
	if err != nil {
		panic(err)
	}

	file, err2 := os.Create(path)
	if err2 != nil {
		panic(err2)
	}
	writer := bufio.NewWriter(file)
	writer.Write(b.Bytes())
	writer.Flush()
}

func PKCS5Padding(ciphertext string, blockSize int) string {
	padding := blockSize - len([]byte(ciphertext))%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return string(append([]byte(ciphertext), padtext...))
}

func PKCS5Trimming(encrypt string) string {
	padding := encrypt[len([]byte(encrypt))-1]
	return string(encrypt[:len(encrypt)-int(padding)])
}

func makeCipher(key string) cipher.Block {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil
	} else {
		return c
	}
}

func encryptAES(c cipher.Block, plaintext string) string {

	// allocate space for ciphered data
	plaintext = PKCS5Padding(plaintext, c.BlockSize())
	out := make([]byte, len(plaintext))

	// encrypt
	ecb := cipher.NewCBCEncrypter(c, []byte(initialVector))
	ecb.CryptBlocks(out, []byte(plaintext))
	// return hex string
	return hex.EncodeToString(out)
}

func decryptAES(c cipher.Block, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)
	ecb := cipher.NewCBCDecrypter(c, []byte(initialVector))
	pt := make([]byte, len(ciphertext))
	ecb.CryptBlocks(pt, ciphertext)
	return PKCS5Trimming(string(pt[:]))
}

func addEntry(site string, user string, pass string) {
	vault_map[site+sep+user] = encryptAES(vault_cipher, pass)
}

func delEntry(site string, user string) {
	delete(vault_map, site+sep+user)
}

func setLoginHandler(a *astilectron.Astilectron) {
	login_w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var s string
		m.Unmarshal(&s)
		if strings.Contains(s, "key"+sep) {
			try.This(func() {
				fmt.Println(s)
				vault_cipher = makeCipher(strings.Split(s, sep)[1])
				loadData(vault_filepath)
				login_w.Hide()
				main_w.Create()
				main_w.Show()
			}).Catch(func(e try.E) {
				time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
				fmt.Println(e)
				showNotification(a, "Incorrect Password", "The password introduced isn't correct")
			})

		}
		return nil
	})
	login_w.On(astilectron.EventNameWindowEventClosed, func(e astilectron.Event) (deleteListener bool) {
		os.Exit(0)
		return
	})
}

func setMainHandler(a *astilectron.Astilectron) {
	main_w.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var s string
		m.Unmarshal(&s)
		if strings.Contains(s, "add"+sep) {
			values := strings.Split(s, sep)
			if len(values[1]) == 0 || len(values[2]) == 0 || len(values[3]) == 0 || strings.Contains(values[1], " ") || strings.Contains(values[2], " ") || strings.Contains(values[3], " ") || strings.Contains(values[1], sep) || strings.Contains(values[2], sep) || strings.Contains(values[3], sep) {
				showNotification(a, "Add Error", "Input values must be free of spaces")
				return nil
			}
			addEntry(values[1], values[2], values[3])
			storeData(vault_map, vault_filepath)
			showNotification(a, values[1], "New Entry")
		} else if strings.Contains(s, "del"+sep) {
			values := strings.Split(strings.Split(s, sep)[1], " @ ")
			delEntry(values[0], values[1])
			storeData(vault_map, vault_filepath)
			showNotification(a, values[0], "Deleted")
		} else if strings.Contains(s, "search"+sep) {
			value := strings.ToLower(strings.Split(s, sep)[1])
			results := ""
			for k := range vault_map {
				if strings.Contains(strings.ToLower(k), value) {
					arr := strings.Split(k, sep)
					results += arr[0] + " @ " + arr[1] + sep
				}
			}
			if last := len(results) - 1; last >= 0 {
				results = results[:last]
			}
			return results
		} else if strings.Contains(s, "get"+sep) {
			value := strings.ReplaceAll(strings.Split(s, sep)[1], " @ ", sep)
			for k, v := range vault_map {
				if k == value {
					return decryptAES(vault_cipher, v)
				}
			}
			return ""
		}
		return nil
	})

	main_w.On(astilectron.EventNameWindowEventClosed, func(e astilectron.Event) (deleteListener bool) {
		os.Exit(0)
		return
	})
}

func closeOnTimeLimit(a *astilectron.Astilectron) {
	time.Sleep(3 * time.Minute)
	showNotification(a, "Vault", "Vault will not longer remain opened due to security reasons")
	time.Sleep(8 * time.Second)
	os.Exit(1)
}

func main() {
	vault_cipher = makeCipher("1234567812345678")
	vault_map = make(map[string]string)

	l := log.New(log.Writer(), log.Prefix(), log.Flags())

	a, err := astilectron.New(l, astilectron.Options{
		AppName:            "Vault",
		AppIconDefaultPath: "app/icon.png",
		AppIconDarwinPath:  "app/icon.icns",
		SingleInstance:     true,
	})
	if err != nil {
		l.Fatal(fmt.Errorf("main: creating astilectron failed: %w", err))
	}
	defer a.Close()

	a.HandleSignals()

	if err = a.Start(); err != nil {
		l.Fatal(fmt.Errorf("main: starting astilectron failed: %w", err))
	}

	if main_w, err = a.NewWindow("./app/main.html", &astilectron.WindowOptions{
		Center:    astikit.BoolPtr(true),
		Height:    astikit.IntPtr(400),
		Width:     astikit.IntPtr(800),
		Resizable: astikit.BoolPtr(false),
	}); err != nil {
		l.Fatal(fmt.Errorf("main: new window failed: %w2", err))
	}

	if login_w, err = a.NewWindow("./app/login.html", &astilectron.WindowOptions{
		Center:    astikit.BoolPtr(true),
		Height:    astikit.IntPtr(350),
		Width:     astikit.IntPtr(500),
		Resizable: astikit.BoolPtr(false),
	}); err != nil {
		l.Fatal(fmt.Errorf("main: new window failed: %w", err))
	}

	if err = login_w.Create(); err != nil {
		l.Fatal(fmt.Errorf("main: creating window failed: %w", err))
	}

	setLoginHandler(a)
	setMainHandler(a)

	go closeOnTimeLimit(a)

	a.Wait()
}
